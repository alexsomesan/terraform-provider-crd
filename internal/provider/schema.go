package provider

import "strings"

var rejectPrefixes = []string{
	"io.k8s.apimachinery.pkg.apis.meta.v1",
	"io.k8s.api.admissionregistration.v1",
}

var rejectSuffixes = []string{"List", "Spec", "Status"}

func rejectPath(p string) bool {
	for _, s := range rejectSuffixes {
		if strings.HasSuffix(p, s) {
			return true
		}
	}
	for _, j := range rejectPrefixes {
		if strings.HasPrefix(p, j) {
			return true
		}
	}
	return false
}

func schemaFromOpenAPI() {

}
