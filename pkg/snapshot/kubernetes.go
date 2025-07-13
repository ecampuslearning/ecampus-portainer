package snapshot

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/segmentio/encoding/json"

	"github.com/aws/smithy-go/ptr"
	portainer "github.com/portainer/portainer/api"
	edgeutils "github.com/portainer/portainer/pkg/edge"
	networkingutils "github.com/portainer/portainer/pkg/networking"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	statsapi "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

func CreateKubernetesSnapshot(cli *kubernetes.Clientset) (*portainer.KubernetesSnapshot, error) {
	kubernetesSnapshot := &portainer.KubernetesSnapshot{}
	err := kubernetesSnapshotVersion(kubernetesSnapshot, cli)
	if err != nil {
		log.Warn().Err(err).Msg("unable to snapshot cluster version")
	}

	err = kubernetesSnapshotNodes(kubernetesSnapshot, cli)
	if err != nil {
		log.Warn().Err(err).Msg("unable to snapshot cluster nodes")
	}

	kubernetesSnapshot.Time = time.Now().Unix()
	return kubernetesSnapshot, nil
}

func kubernetesSnapshotVersion(snapshot *portainer.KubernetesSnapshot, cli *kubernetes.Clientset) error {
	versionInfo, err := cli.ServerVersion()
	if err != nil {
		return err
	}

	snapshot.KubernetesVersion = versionInfo.GitVersion
	return nil
}

func kubernetesSnapshotNodes(snapshot *portainer.KubernetesSnapshot, cli *kubernetes.Clientset) error {
	nodeList, err := cli.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(nodeList.Items) == 0 {
		return nil
	}

	var totalCPUs, totalMemory int64
	performanceMetrics := &portainer.PerformanceMetrics{
		CPUUsage:     0,
		MemoryUsage:  0,
		NetworkUsage: 0,
	}

	for _, node := range nodeList.Items {
		totalCPUs += node.Status.Capacity.Cpu().Value()
		totalMemory += node.Status.Capacity.Memory().Value()

		performanceMetrics, err = kubernetesSnapshotNodePerformanceMetrics(cli, node, performanceMetrics)
		if err != nil {
			return fmt.Errorf("failed to get node performance metrics: %w", err)
		}
		if performanceMetrics != nil {
			snapshot.PerformanceMetrics = performanceMetrics
		}
	}

	snapshot.TotalCPU = totalCPUs
	snapshot.TotalMemory = totalMemory
	snapshot.NodeCount = len(nodeList.Items)
	return nil
}

// KubernetesSnapshotDiagnostics returns the diagnostics data for the agent
func KubernetesSnapshotDiagnostics(cli *kubernetes.Clientset, edgeKey string) (*portainer.DiagnosticsData, error) {
	podID := os.Getenv("HOSTNAME")
	snapshot := &portainer.KubernetesSnapshot{
		DiagnosticsData: &portainer.DiagnosticsData{
			DNS:    make(map[string]string),
			Telnet: make(map[string]string),
		},
	}

	err := kubernetesSnapshotPodErrorLogs(snapshot, cli, "portainer", podID)
	if err != nil {
		return nil, fmt.Errorf("failed to snapshot pod error logs: %w", err)
	}

	if edgeKey != "" {
		url, err := edgeutils.GetPortainerURLFromEdgeKey(edgeKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get portainer URL from edge key: %w", err)
		}

		snapshot.DiagnosticsData.DNS["edge-to-portainer"] = networkingutils.ProbeDNSConnection(url)
		snapshot.DiagnosticsData.Telnet["edge-to-portainer"] = networkingutils.ProbeTelnetConnection(url)
	}

	return snapshot.DiagnosticsData, nil
}

// KubernetesSnapshotPodErrorLogs returns 0 to 10 lines of the most recent error logs of the agent container
// this will primarily be used for agent snapshot
func kubernetesSnapshotPodErrorLogs(snapshot *portainer.KubernetesSnapshot, cli *kubernetes.Clientset, namespace, podID string) error {
	if namespace == "" || podID == "" {
		return errors.New("both namespace and podID are required to capture pod error logs in the snapshot")
	}

	logsStream, err := cli.CoreV1().Pods(namespace).GetLogs(podID, &corev1.PodLogOptions{TailLines: ptr.Int64(10), Timestamps: true}).Stream(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to stream logs: %w", err)
	}
	defer logsStream.Close()

	logBytes, err := io.ReadAll(logsStream)
	if err != nil {
		return fmt.Errorf("failed to read error logs: %w", err)
	}

	logs := filterLogsByPattern(logBytes, []string{"error", "err", "level=error", "exception", "fatal", "panic"})

	jsonLogs, err := json.Marshal(logs)
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}
	snapshot.DiagnosticsData.Log = string(jsonLogs)

	return nil
}

func kubernetesSnapshotNodePerformanceMetrics(cli *kubernetes.Clientset, node corev1.Node, performanceMetrics *portainer.PerformanceMetrics) (*portainer.PerformanceMetrics, error) {
	result := cli.RESTClient().Get().AbsPath(fmt.Sprintf("/api/v1/nodes/%s/proxy/stats/summary", node.Name)).Do(context.TODO())
	if result.Error() != nil {
		return nil, fmt.Errorf("failed to get node performance metrics: %w", result.Error())
	}

	raw, err := result.Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to get node performance metrics: %w", err)
	}

	stats := statsapi.Summary{}
	err = json.Unmarshal(raw, &stats)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal node performance metrics: %w", err)
	}

	nodeStats := stats.Node
	if reflect.DeepEqual(nodeStats, statsapi.NodeStats{}) {
		return nil, nil
	}

	if nodeStats.CPU != nil && nodeStats.CPU.UsageNanoCores != nil {
		performanceMetrics.CPUUsage += math.Round(float64(*nodeStats.CPU.UsageNanoCores) / float64(node.Status.Capacity.Cpu().Value()*1000000000) * 100)
	}
	if nodeStats.Memory != nil && nodeStats.Memory.WorkingSetBytes != nil {
		performanceMetrics.MemoryUsage += math.Round(float64(*nodeStats.Memory.WorkingSetBytes) / float64(node.Status.Capacity.Memory().Value()) * 100)
	}
	if nodeStats.Network != nil && nodeStats.Network.RxBytes != nil && nodeStats.Network.TxBytes != nil {
		performanceMetrics.NetworkUsage += math.Round((float64(*nodeStats.Network.RxBytes) + float64(*nodeStats.Network.TxBytes)) / 1024 / 1024) // MB
	}
	return performanceMetrics, nil
}

// filterLogsByPattern filters the logs by the given patterns and returns a list of logs that match the patterns
// the logs are returned as a list of maps with the keys "timestamp" and "message"
func filterLogsByPattern(logBytes []byte, patterns []string) []map[string]string {
	logs := []map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(logBytes)), "\n") {
		if line == "" {
			continue
		}

		if parts := strings.SplitN(line, " ", 2); len(parts) == 2 {
			messageLower := strings.ToLower(parts[1])
			for _, pattern := range patterns {
				if strings.Contains(messageLower, pattern) {
					logs = append(logs, map[string]string{
						"timestamp": parts[0],
						"message":   parts[1],
					})
					break
				}
			}
		}
	}

	return logs
}
