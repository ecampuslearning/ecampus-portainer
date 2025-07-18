package cli

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/pkg/errors"
	portainer "github.com/portainer/portainer/api"
	models "github.com/portainer/portainer/api/http/models/kubernetes"
	"github.com/portainer/portainer/api/stacks/stackutils"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	systemNamespaceLabel = "io.portainer.kubernetes.namespace.system"
	namespaceOwnerLabel  = "io.portainer.kubernetes.resourcepool.owner"
	namespaceNameLabel   = "io.portainer.kubernetes.resourcepool.name"
)

func defaultSystemNamespaces() map[string]struct{} {
	return map[string]struct{}{
		"kube-system":     {},
		"kube-public":     {},
		"kube-node-lease": {},
		"portainer":       {},
	}
}

// GetNamespaces gets the namespaces in the current k8s environment(endpoint).
// if the user is an admin, all namespaces in the current k8s environment(endpoint) are fetched using the fetchNamespaces function.
// otherwise, namespaces the non-admin user has access to will be used to filter the namespaces based on the allowed namespaces.
func (kcl *KubeClient) GetNamespaces() (map[string]portainer.K8sNamespaceInfo, error) {
	if kcl.IsKubeAdmin {
		return kcl.fetchNamespaces()
	}
	return kcl.fetchNamespacesForNonAdmin()
}

// fetchNamespacesForNonAdmin gets the namespaces in the current k8s environment(endpoint) for the non-admin user.
func (kcl *KubeClient) fetchNamespacesForNonAdmin() (map[string]portainer.K8sNamespaceInfo, error) {
	log.Debug().
		Str("context", "fetchNamespacesForNonAdmin").
		Msg("Fetching namespaces for non-admin user")

	if len(kcl.NonAdminNamespaces) == 0 {
		return nil, nil
	}

	namespaces, err := kcl.fetchNamespaces()
	if err != nil {
		return nil, fmt.Errorf("an error occurred during the fetchNamespacesForNonAdmin operation, unable to list namespaces for the non-admin user: %w", err)
	}

	nonAdminNamespaceSet := kcl.buildNonAdminNamespacesMap()
	results := make(map[string]portainer.K8sNamespaceInfo)
	for _, namespace := range namespaces {
		if _, exists := nonAdminNamespaceSet[namespace.Name]; exists {
			results[namespace.Name] = namespace
		}
	}

	return results, nil
}

// fetchNamespaces gets the namespaces in the current k8s environment(endpoint).
// this function is used by both admin and non-admin users.
// the result gets parsed to a map of namespace name to namespace info.
func (kcl *KubeClient) fetchNamespaces() (map[string]portainer.K8sNamespaceInfo, error) {
	namespaces, err := kcl.cli.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error().
			Str("context", "fetchNamespaces").
			Err(err).
			Msg("Failed to list namespaces")

		return nil, fmt.Errorf("an error occurred during the fetchNamespacesForAdmin operation, unable to list namespaces for the admin user: %w", err)
	}

	results := make(map[string]portainer.K8sNamespaceInfo)
	for _, namespace := range namespaces.Items {
		results[namespace.Name] = parseNamespace(&namespace)
	}

	return results, nil
}

// parseNamespace converts a k8s namespace object to a portainer namespace object.
func parseNamespace(namespace *corev1.Namespace) portainer.K8sNamespaceInfo {
	return portainer.K8sNamespaceInfo{
		Id:             string(namespace.UID),
		Name:           namespace.Name,
		Status:         namespace.Status,
		Annotations:    namespace.Annotations,
		CreationDate:   namespace.CreationTimestamp.Format(time.RFC3339),
		NamespaceOwner: namespace.Labels[namespaceOwnerLabel],
		IsSystem:       isSystemNamespace(namespace),
		IsDefault:      namespace.Name == defaultNamespace,
	}
}

// GetNamespace gets the namespace in the current k8s environment(endpoint).
func (kcl *KubeClient) GetNamespace(name string) (portainer.K8sNamespaceInfo, error) {
	namespace, err := kcl.cli.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Error().
			Str("context", "GetNamespace").
			Str("namespace", name).
			Err(err).
			Msg("Failed to get namespace")
		return portainer.K8sNamespaceInfo{}, err
	}

	return parseNamespace(namespace), nil
}

// CreateNamespace creates a new namespace in a k8s endpoint.
func (kcl *KubeClient) CreateNamespace(info models.K8sNamespaceDetails) (*corev1.Namespace, error) {
	portainerLabels := map[string]string{
		namespaceNameLabel:  stackutils.SanitizeLabel(info.Name),
		namespaceOwnerLabel: stackutils.SanitizeLabel(info.Owner),
	}

	var ns corev1.Namespace
	ns.Name = info.Name
	ns.Annotations = info.Annotations
	ns.Labels = portainerLabels

	namespace, err := kcl.cli.CoreV1().Namespaces().Create(context.Background(), &ns, metav1.CreateOptions{})
	if err != nil {
		log.Error().
			Err(err).
			Str("context", "CreateNamespace").
			Str("Namespace", info.Name).
			Msg("Failed to create the namespace")
		return nil, err
	}

	if err := kcl.createOrUpdateNamespaceResourceQuota(info, portainerLabels); err != nil {
		log.Error().
			Err(err).
			Str("context", "CreateNamespace").
			Str("name", info.Name).
			Msg("failed to create or update resource quota for namespace")
		return nil, err
	}

	return namespace, nil
}

// UpdateIngress updates an ingress in a given namespace in a k8s endpoint.
func (kcl *KubeClient) UpdateNamespace(info models.K8sNamespaceDetails) (*corev1.Namespace, error) {
	portainerLabels := map[string]string{
		namespaceNameLabel:  stackutils.SanitizeLabel(info.Name),
		namespaceOwnerLabel: stackutils.SanitizeLabel(info.Owner),
	}

	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        info.Name,
			Annotations: info.Annotations,
		},
	}

	updatedNamespace, err := kcl.cli.CoreV1().Namespaces().Update(context.Background(), &namespace, metav1.UpdateOptions{})
	if err != nil {
		log.Error().
			Str("context", "UpdateNamespace").
			Str("namespace", info.Name).
			Err(err).
			Msg("Failed to update namespace")
		return nil, err
	}

	if err := kcl.createOrUpdateNamespaceResourceQuota(info, portainerLabels); err != nil {
		log.Error().
			Err(err).
			Str("context", "UpdateNamespace").
			Str("name", info.Name).
			Msg("failed to create or update resource quota for namespace")
		return nil, err
	}

	return updatedNamespace, nil
}

func (kcl *KubeClient) createOrUpdateNamespaceResourceQuota(info models.K8sNamespaceDetails, portainerLabels map[string]string) error {
	if !info.ResourceQuota.Enabled {
		if err := kcl.deleteNamespaceResourceQuota(info.Name); err != nil {
			log.Debug().Err(err).Str("context", "createOrUpdateNamespaceResourceQuota").Str("name", info.Name).Msg("failed to delete resource quota for namespace")
		}
		return nil
	}

	resourceQuota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "portainer-rq-" + info.Name,
			Namespace: info.Name,
			Labels:    portainerLabels,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{},
		},
	}

	if info.ResourceQuota.Enabled {
		memory := resource.MustParse(info.ResourceQuota.Memory)
		cpu := resource.MustParse(info.ResourceQuota.CPU)

		if memory.Value() > 0 {
			memQuota := memory
			resourceQuota.Spec.Hard[corev1.ResourceLimitsMemory] = memQuota
			resourceQuota.Spec.Hard[corev1.ResourceRequestsMemory] = memQuota
		}

		if cpu.Value() > 0 {
			cpuQuota := cpu
			resourceQuota.Spec.Hard[corev1.ResourceLimitsCPU] = cpuQuota
			resourceQuota.Spec.Hard[corev1.ResourceRequestsCPU] = cpuQuota
		}
	}

	_, err := kcl.cli.CoreV1().ResourceQuotas(info.Name).Update(context.Background(), resourceQuota, metav1.UpdateOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Warn().
				Str("context", "createOrUpdateNamespaceResourceQuota").
				Str("name", info.Name).
				Msg("resource quota not found, creating")
			_, err = kcl.cli.CoreV1().ResourceQuotas(info.Name).Create(context.Background(), resourceQuota, metav1.CreateOptions{})
		}
	}

	return err
}

func (kcl *KubeClient) deleteNamespaceResourceQuota(namespaceName string) error {
	err := kcl.cli.CoreV1().ResourceQuotas(namespaceName).Delete(context.Background(), "portainer-rq-"+namespaceName, metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		log.Error().
			Str("context", "deleteNamespaceResourceQuota").
			Str("name", namespaceName).
			Err(err).
			Msg("failed to delete resource quota for namespace")
		return err
	}
	log.Warn().
		Str("context", "deleteNamespaceResourceQuota").
		Str("name", namespaceName).
		Msg("resource quota to delete not found")
	return nil
}

func isSystemNamespace(namespace *corev1.Namespace) bool {
	systemLabelValue, hasSystemLabel := namespace.Labels[systemNamespaceLabel]
	if hasSystemLabel {
		return systemLabelValue == "true"
	}

	return isSystemDefaultNamespace(namespace.Name)
}

func isSystemDefaultNamespace(namespace string) bool {
	systemNamespaces := defaultSystemNamespaces()
	_, isSystem := systemNamespaces[namespace]
	return isSystem
}

func (kcl *KubeClient) isSystemNamespace(namespace string) bool {
	ns, err := kcl.cli.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		return false
	}

	return isSystemNamespace(ns)
}

// ToggleSystemState will set a namespace as a system namespace, or remove this state
// if isSystem is true it will set `systemNamespaceLabel` to "true" and false otherwise
// this will skip if namespace is "default" or if the required state is already set
func (kcl *KubeClient) ToggleSystemState(namespaceName string, isSystem bool) error {
	if namespaceName == "default" {
		return nil
	}

	namespace, err := kcl.cli.CoreV1().Namespaces().Get(context.TODO(), namespaceName, metav1.GetOptions{})
	if err != nil {
		log.Error().
			Str("context", "ToggleSystemState").
			Str("namespace", namespaceName).
			Err(err).
			Msg("failed to get namespace")
		return errors.Wrap(err, "failed fetching namespace object")
	}

	if isSystemNamespace(namespace) == isSystem {
		return nil
	}

	if namespace.Labels == nil {
		namespace.Labels = map[string]string{}
	}

	namespace.Labels[systemNamespaceLabel] = strconv.FormatBool(isSystem)

	if _, err := kcl.cli.CoreV1().Namespaces().Update(context.TODO(), namespace, metav1.UpdateOptions{}); err != nil {
		log.Error().
			Str("context", "ToggleSystemState").
			Str("namespace", namespaceName).
			Err(err).
			Msg("failed updating namespace object")
		return errors.Wrap(err, "failed updating namespace object")
	}

	if isSystem {
		return kcl.NamespaceAccessPoliciesDeleteNamespace(namespaceName)
	}

	return nil
}

func (kcl *KubeClient) DeleteNamespace(namespaceName string) (*corev1.Namespace, error) {
	namespace, err := kcl.cli.CoreV1().Namespaces().Get(context.Background(), namespaceName, metav1.GetOptions{})
	if err != nil {
		log.Error().
			Str("context", "DeleteNamespace").
			Str("namespace", namespaceName).
			Err(err).
			Msg("failed fetching namespace object")
		return nil, err
	}

	err = kcl.cli.CoreV1().Namespaces().Delete(context.Background(), namespaceName, metav1.DeleteOptions{})
	if err != nil {
		log.Error().
			Str("context", "DeleteNamespace").
			Str("namespace", namespaceName).
			Err(err).
			Msg("failed deleting namespace object")
		return nil, err
	}

	return namespace, nil
}

// CombineNamespacesWithUnhealthyEvents combines namespaces with unhealthy events across all namespaces
func (kcl *KubeClient) CombineNamespacesWithUnhealthyEvents(namespaces map[string]portainer.K8sNamespaceInfo) (map[string]portainer.K8sNamespaceInfo, error) {
	allEvents, err := kcl.GetEvents("", "")
	if err != nil && !k8serrors.IsNotFound(err) {
		log.Error().
			Str("context", "CombineNamespacesWithUnhealthyEvents").
			Err(err).
			Msg("unable to retrieve unhealthy events from the Kubernetes for an admin user")
		return nil, err
	}

	unhealthyEventCounts := make(map[string]int)
	for _, event := range allEvents {
		if event.Type == "Warning" {
			unhealthyEventCounts[event.Namespace]++
		}
	}

	for namespaceName, namespace := range namespaces {
		if count, exists := unhealthyEventCounts[namespaceName]; exists {
			namespace.UnhealthyEventCount = count
			namespaces[namespaceName] = namespace
		}
	}

	return namespaces, nil
}

// CombineNamespacesWithResourceQuotas combines namespaces with resource quotas where matching is based on "portainer-rq-"+namespace.Name
func (kcl *KubeClient) CombineNamespacesWithResourceQuotas(namespaces map[string]portainer.K8sNamespaceInfo, w http.ResponseWriter) *httperror.HandlerError {
	resourceQuotas, err := kcl.GetResourceQuotas("")
	if err != nil && !k8serrors.IsNotFound(err) {
		log.Error().
			Str("context", "CombineNamespacesWithResourceQuotas").
			Err(err).
			Msg("unable to retrieve resource quotas from the Kubernetes for an admin user")
		return httperror.InternalServerError("an error occurred during the CombineNamespacesWithResourceQuotas operation, unable to retrieve resource quotas from the Kubernetes for an admin user. Error: ", err)
	}

	if len(*resourceQuotas) > 0 {
		return response.JSON(w, kcl.UpdateNamespacesWithResourceQuotas(namespaces, *resourceQuotas))
	}

	return response.JSON(w, kcl.ConvertNamespaceMapToSlice(namespaces))
}

// CombineNamespaceWithResourceQuota combines a namespace with a resource quota prefixed with "portainer-rq-"+namespace.Name
func (kcl *KubeClient) CombineNamespaceWithResourceQuota(namespace portainer.K8sNamespaceInfo, w http.ResponseWriter) *httperror.HandlerError {
	resourceQuota, err := kcl.GetPortainerResourceQuota(namespace.Name)
	if err != nil && !k8serrors.IsNotFound(err) {
		log.Error().
			Str("context", "CombineNamespaceWithResourceQuota").
			Str("namespace", namespace.Name).
			Err(err).
			Msg("unable to retrieve the resource quota associated with the namespace")
		return httperror.InternalServerError(fmt.Sprintf("an error occurred during the CombineNamespaceWithResourceQuota operation, unable to retrieve the resource quota associated with the namespace: %s for a non-admin user. Error: ", namespace.Name), err)
	}

	if resourceQuota != nil {
		namespace.ResourceQuota = resourceQuota
	}

	return response.JSON(w, namespace)
}

// buildNonAdminNamespacesMap builds a map of non-admin namespaces.
// the map is used to filter the namespaces based on the allowed namespaces.
func (kcl *KubeClient) buildNonAdminNamespacesMap() map[string]struct{} {
	nonAdminNamespaceSet := make(map[string]struct{}, len(kcl.NonAdminNamespaces))
	for _, namespace := range kcl.NonAdminNamespaces {
		if !isSystemDefaultNamespace(namespace) {
			nonAdminNamespaceSet[namespace] = struct{}{}
		}
	}

	return nonAdminNamespaceSet
}

// ConvertNamespaceMapToSlice converts the namespace map to a slice of namespaces.
// this is used to for the API response.
func (kcl *KubeClient) ConvertNamespaceMapToSlice(namespaces map[string]portainer.K8sNamespaceInfo) []portainer.K8sNamespaceInfo {
	namespaceSlice := make([]portainer.K8sNamespaceInfo, 0, len(namespaces))
	for _, namespace := range namespaces {
		namespaceSlice = append(namespaceSlice, namespace)
	}

	// Sort namespaces by name
	sort.Slice(namespaceSlice, func(i, j int) bool {
		return namespaceSlice[i].Name < namespaceSlice[j].Name
	})

	return namespaceSlice
}
