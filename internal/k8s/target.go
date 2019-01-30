package k8s

import (
	"fmt"

	"github.com/windmilleng/tilt/internal/model"
)

func NewTarget(name model.TargetName, entities []K8sEntity, portForwards []model.PortForward) (model.K8sTarget, error) {
	yaml, err := SerializeYAML(entities)
	if err != nil {
		return model.K8sTarget{}, err
	}

	var resourceNames []string
	for _, e := range entities {
		resourceNames = append(resourceNames, resourceName(e))
	}

	return model.K8sTarget{
		Name:          name,
		YAML:          yaml,
		ResourceNames: resourceNames,
		PortForwards:  portForwards,
	}, nil
}

func NewK8sOnlyManifest(name model.ManifestName, entities []K8sEntity) (model.Manifest, error) {
	kTarget, err := NewTarget(name.TargetName(), entities, nil)
	if err != nil {
		return model.Manifest{}, err
	}
	return model.Manifest{Name: name}.WithDeployTarget(kTarget), nil
}

func NewK8sOnlyManifestsPerEntity(entities []K8sEntity) ([]model.Manifest, error) {
	manifests := make([]model.Manifest, len(entities))
	for i, e := range entities {
		name := model.ManifestName(resourceName(e))

		m, err := NewK8sOnlyManifest(name, []K8sEntity{e})
		if err != nil {
			return nil, err
		}
		manifests[i] = m
	}
	return manifests, nil
}

func NewK8sOnlyManifestForTesting(name model.ManifestName, yaml string) model.Manifest {
	return model.Manifest{Name: name}.
		WithDeployTarget(model.K8sTarget{Name: name.TargetName(), YAML: yaml})
}

func resourceName(e K8sEntity) string {
	return fmt.Sprintf("k8s%s-%s", e.Kind.Kind, e.Name())
}
