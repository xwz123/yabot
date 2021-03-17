# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

load("@io_bazel_rules_docker//container:image.bzl", "container_image")
load("@io_bazel_rules_docker//container:bundle.bzl", "container_bundle")
load("@io_bazel_rules_docker//contrib:push-all.bzl", "container_push")
load("@io_bazel_rules_docker//go:image.bzl", "go_image")
load(
     "//:image.bzl",
     _image_tags = "tags",
)

## prow_image is a macro for creating :app and :image targets
def build_image(
        name,  # use "image"
        base = None,
        stamp = True,  # stamp by default, but allow overrides
        app_name = "app",
        **kwargs):
    go_image(
        name = app_name,
        base = base,
        embed = [":go_default_library"],
        goarch = "amd64",
        goos = "linux",
    )

    container_image(
        name = name,
        base = ":" + app_name,
        stamp = stamp,
        **kwargs
    )

# push_image creates a bundle of container images, and a target to push them.
def push_image(
        name,
        bundle_name = "bundle",
        images = None):
    container_bundle(
        name = bundle_name,
        images = images,
    )
    container_push(
        name = name,
        bundle = ":" + bundle_name,
        format = "Docker",
    )

# prefix returns the image prefix for the command.
#
# Concretely, image("foo") returns "{EDGE_PROW_REPO}/foo"
# which usually becomes xwzqmxx/foo
def prefix(cmd):
    return "{STABLE_PROW_REPO}%s" % cmd

# tags returns a {image: target} map for each cmd or {name: target}
# Concretely, tags("hook",  **{"ghproxy": "//ghproxy:image"}) will output the following:
#  {
#   "xwzqmxx/hook:20210203-defada"://prow/cmd/hook:image
#   "xwzqmxx/hook:latest" //prow/cmd/hook:image
#   "xwzqmxx/hook:latest-root" //prow/cmd/hook:image
#   "xwzqmxx/ghproxy:20180203-deadbeef": "//ghproxy:image",
#   "xwzqmxx/ghproxy:latest": "//ghproxy:image",
#   "xwzqmxx/ghproxy:latest-fejta": "//ghproxy:image",
#  }

def tags(cmds, targets):
    cmd_targets = {prefix(cmd): target(cmd) for cmd in cmds}
    return _image_tags(cmd_targets)