#!/usr/bin/env bash

export PROJECT_ID=bobsknobshop
export GOOGLE_CLOUD_PROJECT=bobsknobshop

# Project Variables
export PROJECT_ORGANISATION=bobsknobshop

# Expected to be set/overridden in personal-envs.sh
# Placed here for reference, DO NOT EDIT
GOOGLE_ACCOUNT=
GOOGLE_APPLICATION_CREDENTIALS=
PROTO_GOOGLE_APIS=
GENPROTO_REPO=

# Default executables, override in personal-envs.sh as needed.
PYTHON27_EXEC="$(which python2.7)"
PYTHON35P_EXEC="$(which python3.6)"
VENV_EXEC="$(which virtualenv)"
PROTOC_EXEC="$(which protoc)"
DOCKER="$(which docker)"
MAKE="$(which maker)"


if [[ -f personal-envs.sh ]]; then
    . personal-envs.sh
fi

if [[ -z ${GOOGLE_ACCOUNT} ]]; then
    echo "Error: Update personal-envs.sh with your personal settings first."
else
    if [[ -z "${PROJECT_ID}" || \
            "${PROJECT_ID}" = "PLACEHOLDER_*" ]]; then
        echo "Error: PROJECT_ID not set"
    else
        gcloud config set account ${GOOGLE_ACCOUNT} &> /dev/null
        gcloud config set project ${PROJECT_ID} &> /dev/null
    fi
fi
