#!/usr/bin/env bash

function initVanillaController(){
  VANILLA_USER=$(echo "$VANILLA_CONTROLLER" | sed "s|@.*||g")
  VANILLA_HOST=$(echo "$VANILLA_CONTROLLER" | sed "s|.*@||g")
  VANILLA_PORT=$(echo "$VANILLA_CONTROLLER" | cut -d':' -s -f2)
  VANILLA_PORT="${PORT:-22}"
}

function initAllLocalDeleteFile() {
  cat test/conf/local.yaml > test/conf/all-local.yaml
  echo "" >> test/conf/all-local.yaml
  cat test/conf/local-agent.yaml >> test/conf/all-local.yaml
  echo "" >> test/conf/all-local.yaml
  cat test/conf/application.yaml >> test/conf/all-local.yaml
}

function initMicroserviceFile() {
  echo "---
apiVersion: iofog.org/v1
kind: Microservice 
metadata:
  name: ${MICROSERVICE_NAME}
spec:
  agent:
    name: ${NAME}-0
    config:
      memoryLimit: 8192
  images:
    arm: edgeworx/healthcare-heart-rate:test-arm
    x86: edgeworx/healthcare-heart-rate:test
    registry: remote # public docker
  container:
    rootHostAccess: false
    volumes:
      - hostDestination: /tmp/microservice
        containerDestination: /tmp
        accessMode: rw
    ports:
      - internal: 443
        external: 5005
    env:
      - key: TEST
        value: 42
  application: ${APPLICATION_NAME}
  routes:
    - ${MSVC1_NAME}
    - ${MSVC2_NAME}
  config:
    test_mode: true
    data_label: 'Anonymous_Person_2'" > test/conf/microservice.yaml
}

function initMicroserviceUpdateFile() {
  echo "---
apiVersion: iofog.org/v1
kind: Microservice
metadata:
  name: ${MICROSERVICE_NAME}
spec:
  agent:
    name: ${NAME}-0
    config:
      memoryLimit: 5555
      diskDirectory: /tmp/iofog-agent/
  images:
    arm: edgeworx/healthcare-heart-rate:test-arm
    x86: edgeworx/healthcare-heart-rate:test
    registry: remote # public docker
  container:
    rootHostAccess: false
    volumes:
      - hostDestination: /tmp/updatedmicroservice
        containerDestination: /tmp
        accessMode: rw
    ports:
      - internal: 443
        external: 5443
      - internal: 80
        external: 5080
    env:
      - key: TEST
        value: 75
      - key: TEST_2
        value: 42
  application: ${APPLICATION_NAME}
  routes:
    - ${MSVC1_NAME}
  config:
    test_mode: true
    test_data:
      key: 42
    data_label: 'Anonymous_Person_3'" > test/conf/updatedMicroservice.yaml
}

function initApplicationFiles() {
  MSVCS="
    microservices:
    - name: $MSVC1_NAME
      agent:
        name: ${NAME}-0
        config:
          bluetoothEnabled: true # this will install the iofog/restblue microservice
          abstractedHardwareEnabled: false
      images:
        arm: edgeworx/healthcare-heart-rate:arm-v1
        x86: edgeworx/healthcare-heart-rate:x86-v1
        registry: remote # public docker
      container:
        rootHostAccess: false
        volumes:
          - hostDestination: /tmp/msvc
            containerDestination: /tmp
            accessMode: z
        ports: []
      config:
        test_mode: true
        data_label: 'Anonymous_Person'
    # Simple JSON viewer for the heart rate output
    - name: $MSVC2_NAME
      agent:
        name: ${NAME}-0
      images:
        arm: edgeworx/healthcare-heart-rate-ui:arm
        x86: edgeworx/healthcare-heart-rate-ui:x86
        registry: remote
      container:
        rootHostAccess: false
        ports:
          # The ui will be listening on port 80 (internal).
          - external: 5000
            internal: 80
            public: 5000
        volumes: []
        env:
          - key: BASE_URL
            value: http://localhost:8080/data"
  ROUTES="
    routes:
    # Use this section to configure route between microservices
    # Use microservice name
    - from: $MSVC1_NAME
      to: $MSVC2_NAME"

  echo -n "---
  apiVersion: iofog.org/v1
  kind: Application
  metadata:
    name: $APPLICATION_NAME
  spec:" > test/conf/application.yaml
  echo -n "$MSVCS" >> test/conf/application.yaml
  echo "$ROUTES" >> test/conf/application.yaml
}

function initLocalAgentFile() {
  echo "---
apiVersion: iofog.org/v1
kind: Agent
metadata:
  name: ${NAME}-0
spec:
  host: 127.0.0.1
  container:
    image: ${AGENT_IMAGE}" > test/conf/local-agent.yaml
}

function initLocalControllerFile() {
    echo "---
apiVersion: iofog.org/v1
kind: ControlPlane
spec:
  iofogUser:
    name: Testing
    surname: Functional
    email: user@domain.com
    password: S5gYVgLEZV
  controllers:
  - name: $NAME
    host: 127.0.0.1
    container:
      image: ${CONTROLLER_IMAGE}"> test/conf/local.yaml
}

function initAgentsFile() {
  initAgents
  # Empty file
  echo -n "" > test/conf/agents.yaml
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    echo "---
apiVersion: iofog.org/v1
kind: Agent 
metadata:
  name: $AGENT_NAME
spec:
  host: ${HOSTS[$IDX]}
  ssh:
    user: ${USERS[$IDX]}
    keyFile: $KEY_FILE
  package:
    repo: $AGENT_REPO
    version: $AGENT_VANILLA_VERSION
    token: $AGENT_PACKAGE_CLOUD_TOKEN" >> test/conf/agents.yaml

  echo "====> Agent File:"
  cat test/conf/agents.yaml
  done
}

function initAgents(){
  USERS=()
  HOSTS=()
  PORTS=()
  AGENT_NAMES=()
  AGENTS=($AGENT_LIST)
  for AGENT in "${AGENTS[@]}"; do
    local USER=$(echo "$AGENT" | sed "s|@.*||g")
    local HOST=$(echo "$AGENT" | sed "s|.*@||g")
    local PORT=$(echo "$AGENT" | cut -d':' -s -f2)
    local PORT="${PORT:-22}"

    USERS+=" "
    USERS+="$USER"
    HOSTS+=" "
    HOSTS+="$HOST"
    PORTS+=" "
    PORTS+="$PORT"
    AGENT_NAMES+=" "
    AGENT_NAMES+="$AGENT_NAME"
  done
  USERS=($USERS)
  HOSTS=($HOSTS)
  PORTS=($PORTS)
}

function initGCRRegistryFile() {
  echo "---
kind: Registry
apiVersion: iofog.org/v1
spec:
  url: gcr.io
  email: alex@edgeworx.io
  username: _json_key
  password: my_fake_password
  private: true
  " > test/conf/gcr.yaml
}

function initUpdatedGCRRegistryFile() {
  echo "---
kind: Registry
apiVersion: iofog.org/v1
spec:
  id: 3
  url: https://gcr.io
  email: alex@edgeworx.io
  username: _json_key
  password: my_fake_password
  private: true
  " > test/conf/gcr.yaml
}