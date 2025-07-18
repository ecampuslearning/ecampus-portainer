package images

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/docker/docker/api/types"

	"github.com/containers/image/v5/docker/reference"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

type ImageID string

// Image holds information about an image.
type Image struct {
	// Domain is the registry host of this image
	Domain string
	// Path may include username like portainer/portainer-ee, no Tag or Digest
	Path    string
	Tag     string
	Digest  digest.Digest
	HubLink string
	named   reference.Named
	Opts    ParseImageOptions `json:"-"`
}

// ParseImageOptions holds image options for parsing.
type ParseImageOptions struct {
	Name   string
	HubTpl string
}

// Name returns the full name representation of an image but no Tag or Digest.
func (i *Image) Name() string {
	return i.named.Name()
}

// FullName return the real full name may include Tag or Digest of the image, Tag first.
func (i *Image) FullName() string {
	if i.Tag == "" {
		return i.Name() + "@" + i.Digest.String()
	}

	return i.Name() + ":" + i.Tag
}

// String returns the string representation of an image, including Tag and Digest if existed.
func (i *Image) String() string {
	return i.named.String()
}

// Reference returns either the digest if it is non-empty or the tag for the image.
func (i *Image) Reference() string {
	if len(i.Digest.String()) > 1 {
		return i.Digest.String()
	}

	return i.Tag
}

// WithDigest sets the digest for an image.
func (i *Image) WithDigest(digest digest.Digest) (err error) {
	i.Digest = digest
	i.named, err = reference.WithDigest(i.named, digest)

	return err
}

func (i *Image) WithTag(tag string) (err error) {
	i.Tag = tag
	i.named, err = reference.WithTag(i.named, tag)

	return err
}

func (i *Image) TrimDigest() error {
	i.Digest = ""
	named, err := ParseImage(ParseImageOptions{Name: i.FullName()})
	if err != nil {
		return err
	}
	i.named = &named

	return nil
}

// ParseImage returns an Image struct with all the values filled in for a given image.
func ParseImage(parseOpts ParseImageOptions) (Image, error) {
	// Parse the image name and tag.
	named, err := reference.ParseNormalizedNamed(parseOpts.Name)
	if err != nil {
		return Image{}, errors.Wrapf(err, "parsing image %s failed", parseOpts.Name)
	}

	// Add the latest lag if they did not provide one.
	named = reference.TagNameOnly(named)

	i := Image{
		Opts:   parseOpts,
		named:  named,
		Domain: reference.Domain(named),
		Path:   reference.Path(named),
	}

	// Hub link
	i.HubLink, err = i.hubLink()
	if err != nil {
		return Image{}, errors.Wrap(err, fmt.Sprintf("resolving hub link for image %s failed", parseOpts.Name))
	}

	// Add the tag if there was one.
	if tagged, ok := named.(reference.Tagged); ok {
		i.Tag = tagged.Tag()
	}

	// Add the digest if there was one.
	if canonical, ok := named.(reference.Canonical); ok {
		i.Digest = canonical.Digest()
	}

	return i, nil
}

func (i *Image) hubLink() (string, error) {
	if i.Opts.HubTpl != "" {
		var out bytes.Buffer
		tmpl, err := template.New("tmpl").
			Option("missingkey=error").
			Parse(i.Opts.HubTpl)
		if err != nil {
			return "", err
		}
		err = tmpl.Execute(&out, i)

		return out.String(), err
	}

	switch i.Domain {
	case "docker.io":
		prefix := "r"
		path := i.Path
		if strings.HasPrefix(i.Path, "library/") {
			prefix = "_"
			path = strings.Replace(i.Path, "library/", "", 1)
		}

		return "https://hub.docker.com/" + prefix + "/" + path, nil
	case "docker.bintray.io", "jfrog-docker-reg2.bintray.io":
		return "https://bintray.com/jfrog/reg2/" + strings.ReplaceAll(i.Path, "/", "%3A"), nil
	case "docker.pkg.github.com":
		return "https://github.com/" + filepath.ToSlash(filepath.Dir(i.Path)) + "/packages", nil
	case "gcr.io":
		return "https://" + i.Domain + "/" + i.Path, nil
	case "ghcr.io":
		ref := strings.Split(i.Path, "/")
		ghUser, ghPackage := ref[0], ref[1]
		return "https://github.com/users/" + ghUser + "/packages/container/package/" + ghPackage, nil
	case "quay.io":
		return "https://quay.io/repository/" + i.Path, nil
	case "registry.access.redhat.com":
		return "https://access.redhat.com/containers/#/registry.access.redhat.com/" + i.Path, nil
	case "registry.gitlab.com":
		return "https://gitlab.com/" + i.Path + "/container_registry", nil
	default:
		return "", nil
	}
}

// IsLocalImage checks if the image has been built locally
func IsLocalImage(image types.ImageInspect) bool {
	return len(image.RepoDigests) == 0
}

// IsDanglingImage returns whether the given image is "dangling" which means
// that there are no repository references to the given image and it has no
// child images
func IsDanglingImage(image types.ImageInspect) bool {
	return len(image.RepoTags) == 1 && image.RepoTags[0] == "<none>:<none>" && len(image.RepoDigests) == 1 && image.RepoDigests[0] == "<none>@<none>"
}

// IsNoTagImage returns whether the given image is damaged, has no tags
func IsNoTagImage(image types.ImageInspect) bool {
	return len(image.RepoTags) == 0
}
