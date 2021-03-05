module github.com/opensourceways/yabot

go 1.13

replace github.com/golang/lint => golang.org/x/lint v0.0.0-20190301231843-5614ed5bae6f

// Pin all k8s.io staging repositories to kubernetes v0.17.3
// When bumping Kubernetes dependencies, you should update each of these lines
// to point to the same kubernetes v0.KubernetesMinor.KubernetesPatch version
// before running update-deps.sh.
replace (
	cloud.google.com/go => cloud.google.com/go v0.44.3
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	golang.org/x/lint => golang.org/x/lint v0.0.0-20190409202823-959b441ac422
	k8s.io/api => k8s.io/api v0.17.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/client-go => k8s.io/client-go v0.17.3
	k8s.io/code-generator => k8s.io/code-generator v0.17.3
)

require (
	gitee.com/openeuler/go-gitee v0.0.0-20210226091009-de349c8d2916
	github.com/antihax/optional v1.0.0
	github.com/huaweicloud/golangsdk v0.0.0-20210302113304-41351a12edfc
	github.com/prometheus/client_golang v1.5.0
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	k8s.io/apimachinery v0.17.3
	k8s.io/test-infra v0.0.0-20200522021239-7ab687ff3213
	sigs.k8s.io/yaml v1.2.0
)
