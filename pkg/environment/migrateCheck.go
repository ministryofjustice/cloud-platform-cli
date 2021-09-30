package environment

import (
	"github.com/gookit/color"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/ingress"
)

// MigrateCheck check the namespace has the correct externalDNS annotations
func MigrateCheck(ns string) error {

	if ns == "" {
		re := RepoEnvironment{}

		// this already checks we are within the environment repo.
		err := re.mustBeInANamespaceFolder()
		if err != nil {
			return err
		}

		ns, err = re.getNamespaceName()
		if err != nil {
			return err
		}
	}

	hasAnn, err := hasExternalDNSAnnotations(ns)
	if err != nil {
		return err
	}

	if hasAnn == false {
		color.Error.Printf("Namespace: %s doesn't have the correct ingress annotation.\n", ns)
	}

	return nil
}

func hasExternalDNSAnnotations(ns string) (bool, error) {
	var host string = "https://reports.cloud-platform.service.justice.gov.uk/ingress_weighting"

	data, err := ingress.CheckAnnotation(&host)
	if err != nil {
		return false, err
	}

	for _, ingress := range data.WeightingIngress {
		if ingress.Namespace == ns {
			return false, nil
		}
	}

	return true, nil
}