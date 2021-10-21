#!/bin/bash
vendor/k8s.io/code-generator/generate-groups.sh all \
	github.com/raihankhan/httpApiServer-controller/pkg/client \
	github.com/raihankhan/httpApiServer-controller/pkg/apis \
	raihankhan.github.io:v1alpha1
	#--go-header-file /home/office/go/src/github.com/Tasdidur/xcrd/vendor/k8s.io/code-generator/hack/boilerplate.go.txt