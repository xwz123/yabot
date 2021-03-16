load("@io_bazel_rules_docker//container:pull.bzl", "container_pull")

def containers():
    container_pull(
        name = "git-base",
        digest = "sha256:01b0f83fe91b782ec7ddf1e742ab7cc9a2261894fd9ab0760ebfd39af2d6ab28",  # 2018/07/02
        registry = "gcr.io",
        repository = "k8s-prow/git",
        tag = "0.2",
    )
