#!/bin/bash
EXCLUSIONS=(operator.yaml) hack/deploy-setup.sh ${NAMESPACE}
source hack/env.sh
go run ./main.go
