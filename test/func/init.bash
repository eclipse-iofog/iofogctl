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
apiVersion: iofog.org/v3
kind: Microservice 
metadata:
  name: ${APPLICATION_NAME}/${MICROSERVICE_NAME}
spec:
  agent:
    name: ${NAME}-0
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
  config:
    test_mode: true
    data_label: 'Anonymous_Person_2'" > test/conf/microservice.yaml
}

function initRouteFile() {
  echo "---
apiVersion: iofog.org/v3
kind: Route
metadata:
  name: $APPLICATION_NAME/$ROUTE_NAME
spec:
  from: $MSVC1_NAME
  to: $MSVC2_NAME" > test/conf/route.yaml
}

function initMicroserviceUpdateFile() {
  echo "---
apiVersion: iofog.org/v3
kind: Microservice
metadata:
  name: ${APPLICATION_NAME}/${MICROSERVICE_NAME}
spec:
  agent:
    name: ${NAME}-0
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
  config:
    test_mode: true
    test_data:
      key: 42
    data_label: 'Anonymous_Person_3'" > test/conf/updatedMicroservice.yaml
}

function initApplicationFileWithRoutes() {
  initApplicationFiles
  ROUTES="
    routes:
    - name: $ROUTE_NAME
      from: $MSVC1_NAME
      to: $MSVC2_NAME"
  echo "$ROUTES" >> test/conf/application.yaml
}

function initApplicationWithPortFile() {
  MSVCS="
    microservices:
    - name: $MSVC1_NAME
      agent:
        name: ${NAME}-0
      images:
        arm: edgeworx/healthcare-heart-rate:arm-v1
        x86: edgeworx/healthcare-heart-rate:x86-v1
        registry: remote # public docker
      container:
        rootHostAccess: false
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
            public:
              schemes:
              - http
              protocol: http
              router:
                port: 6666
        volumes:
        - hostDestination: $VOL_DEST
          containerDestination: $VOL_CONT_DEST
          accessMode: rw
        env:
          - key: BASE_URL
            value: http://localhost:8080/data"
  echo -n "---
  apiVersion: iofog.org/v3
  kind: Application
  metadata:
    name: $APPLICATION_NAME
  spec:" > test/conf/application.yaml
  echo -n "$MSVCS" >> test/conf/application.yaml
}

function initApplicationWithoutPortsFiles() {
  MSVCS="
    microservices:
    - name: $MSVC1_NAME
      agent:
        name: ${NAME}-0
      images:
        arm: edgeworx/healthcare-heart-rate:arm-v1
        x86: edgeworx/healthcare-heart-rate:x86-v1
        registry: remote # public docker
      container:
        rootHostAccess: false
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
        volumes:
        - hostDestination: $VOL_DEST
          containerDestination: $VOL_CONT_DEST
          accessMode: rw
        env:
          - key: BASE_URL
            value: http://localhost:8080/data"
  echo -n "---
  apiVersion: iofog.org/v3
  kind: Application
  metadata:
    name: $APPLICATION_NAME
  spec:" > test/conf/application.yaml
  echo -n "$MSVCS" >> test/conf/application.yaml
}

function initInvalidApplicationFiles() {
    MSVCS="
    microservices:
    - name: $MSVC3_NAME
      agent:
        name: ${NAME}-0
      images:
        arm: edgeworx/healthcare-heart-rate:arm-v1
        x86: edgeworx/healthcare-heart-rate:x86-v1
        registry: remote # public docker
      container:
        rootHostAccess: false
        ports: []
      config:
        test_mode: true
        data_label: 'Anonymous_Person'
    # Simple JSON viewer for the heart rate output
    - name: $MSVC4_NAME
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
          - external: 5001
            internal: 81
            public:
              schemes:
              - http
              protocol: http
        volumes:
        - hostDestination: $VOL_INVALID_DEST
          containerDestination: $VOL_CONT_INVALID_DEST
          accessMode: rw
        env:
          - key: BASE_URL
            value: http://localhost:8080/data"
  echo -n "---
  apiVersion: iofog.org/v3
  kind: Application
  metadata:
    name: ${APPLICATION_NAME}-0
  spec:" > test/conf/application_volume_missing.yaml
  echo -n "$MSVCS" >> test/conf/application_volume_missing.yaml
}

function initDockerPullStatsApplicationFiles() {
  MSVCS="
    microservices:
    - name: $MSVC5_NAME
      agent:
        name: ${NAME}-0
      images:
        x86: edgeworx/thermal-edge-ai-2-arm:2.0.2
        registry: remote # public docker
      container:
        rootHostAccess: false
        ports: []
      config:
        test_mode: true
        data_label: 'Anonymous_Person'"
  echo -n "---
  apiVersion: iofog.org/v3
  kind: Application
  metadata:
    name: ${APPLICATION_NAME}-1
  spec:" > test/conf/application_pull_stat.yaml
  echo -n "$MSVCS" >> test/conf/application_pull_stat.yaml
}

function initApplicationFiles() {
  MSVCS="
    microservices:
    - name: $MSVC1_NAME
      agent:
        name: ${NAME}-0
      images:
        arm: edgeworx/healthcare-heart-rate:arm-v1
        x86: edgeworx/healthcare-heart-rate:x86-v1
        registry: remote # public docker
      container:
        rootHostAccess: false
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
            public:
              schemes:
              - http
              protocol: http
        volumes:
        - hostDestination: $VOL_DEST
          containerDestination: $VOL_CONT_DEST
          accessMode: rw
        env:
          - key: BASE_URL
            value: http://localhost:8080/data"
  echo -n "---
  apiVersion: iofog.org/v3
  kind: Application
  metadata:
    name: $APPLICATION_NAME
  spec:" > test/conf/application.yaml
  echo -n "$MSVCS" >> test/conf/application.yaml
}

function initLocalAgentFile() {
  echo "---
apiVersion: iofog.org/v3
kind: LocalAgent
metadata:
  name: ${NAME}-0
spec:
  config:
    bluetoothEnabled: true
    abstractedHardwareEnabled: false
    memoryLimit: 8192
    diskDirectory: /tmp/iofog-agent/
    description: special test agent
    latitude: 46.464646
    longitude: 64.646464
    diskLimit: 77
    cpuLimit: 89
    logLimit: 12
    logFileCount: 11
    statusFrequency: 9
    changeFrequency: 8
    deviceScanFrequency: 61
  container:
    image: ${AGENT_IMAGE}" > test/conf/local-agent.yaml
}

function initLocalControllerFile() {
    echo "---
apiVersion: iofog.org/v3
kind: LocalControlPlane
spec:
  iofogUser:
    name: Testing
    surname: Functional
    email: $USER_EMAIL
    password: $USER_PW
  controller:
    name: $NAME
    container:
      image: ${CONTROLLER_IMAGE}"> test/conf/local.yaml
}

function initRemoteAgentsFile() {
  initAgents
  # Empty file
  echo -n "" > test/conf/agents.yaml
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    echo "---
apiVersion: iofog.org/v3
kind: Agent
metadata:
  name: $AGENT_NAME
spec:
  config:
    bluetoothEnabled: true
    abstractedHardwareEnabled: false
    memoryLimit: 8192
    diskDirectory: /tmp/iofog-agent/
    description: special test agent
    latitude: 46.464646
    longitude: 64.646464
    diskLimit: 77
    cpuLimit: 89
    logLimit: 12
    logFileCount: 11
    statusFrequency: 9
    changeFrequency: 8
    deviceScanFrequency: 61
  host: ${HOSTS[$IDX]}
  ssh:
    user: ${USERS[$IDX]}
    keyFile: $KEY_FILE" >> test/conf/agents.yaml
    # Pairs of Agents, one is regular, other is custom
    if [ $(($IDX % 2)) -eq 0 ]; then
      echo "
  package:
    repo: $AGENT_REPO
    version: $AGENT_VANILLA_VERSION
    token: $AGENT_PACKAGE_CLOUD_TOKEN" >> test/conf/agents.yaml
    else
      echo "
  scripts:
    dir: assets/agent
    deps:
      entrypoint: install_deps.sh
    install:
      entrypoint: install_iofog.sh
      args:
      - $AGENT_VANILLA_VERSION
      - $AGENT_REPO
      - $AGENT_PACKAGE_CLOUD_TOKEN
    uninstall:
      entrypoint: uninstall_iofog.sh" >> test/conf/agents.yaml
    fi
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
apiVersion: iofog.org/v3
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
apiVersion: iofog.org/v3
spec:
  id: 3
  url: https://gcr.io
  email: alex@edgeworx.io
  username: _json_key
  password: my_fake_password
  private: true
  " > test/conf/gcr.yaml
}

function initEdgeResourceFile() {
  local ER_VERSION="$EDGE_RESOURCE_VERSION"
  if [ ! -z "$1" ]; then
    ER_VERSION="$1"
  fi
  echo "---
apiVersion: iofog.org/v3
kind: EdgeResource
metadata:
  name: $EDGE_RESOURCE_NAME
spec:
  version: $ER_VERSION
  description: $EDGE_RESOURCE_DESC
  interfaceProtocol: $EDGE_RESOURCE_PROTOCOL
  orchestrationTags:
  - smart
  - door
  interface:
    endpoints:
    - name: open
      method: PUT
      url: '/open'
    - name: close
      method: PUT
      url: '/close'
  display:
    name: 'Smart Door'
    icon: 'icon'
    color: '#fefefefe'" > test/conf/edge-resource.yaml
}

function initApplicationTemplateFile(){
  echo -n "---
apiVersion: iofog.org/v3
kind: Application
metadata:
  name: $APPLICATION_NAME
spec:
  template:
    name: $APP_TEMPLATE_NAME
    variables:
    - key: root-host-access
      value: true
    - key: rebuild
      value: false
    - key: magic-number
      value: 12345
    - key: internal
      value: 80
    - key: external
      value: 7777
    - key: turtle
      value:
        turtles:
        - name: john
          job: johnb
          age: 139
          info:
            likes:
            - cats
            - hats
        - name: bob
          job: bobbing
          age: 121
          info:
            lifes:
            - pineapple
    - key: $APP_TEMPLATE_KEY
      value: $APP_TEMPLATE_DEF_VAL" > test/conf/templated-app.yaml
  echo "---
apiVersion: iofog.org/v3
kind: ApplicationTemplate
metadata:
  name: $APP_TEMPLATE_NAME
spec:
    name: $APP_TEMPLATE_NAME
    description: $APP_TEMPLATE_DESC
    variables:
    - key: $APP_TEMPLATE_KEY
      description: $APP_TEMPLATE_KEY_DESC
      defaultValue: $APP_TEMPLATE_DEF_VAL
    - key: rebuild
      description: custom
      defaultValue: true
    - key: root-host-access
      description: custom
      defaultValue: false
    - key: magic-number
      description: custom
      defaultValue: 123
    - key: public-port
      defaultValue: 6666
    - key: turtle
      description: custom
      defaultValue:
        turtles:
        - name: peter
          job: peteing
          age: 100
          info:
            likes:
            - toes
            - shoes
        - name: bob
          job: bobbing
          age: 101
          info:
            lifes:
            - pineapple
    application:
      routes:
      - name: $ROUTE_NAME
        from: $MSVC1_NAME
        to: $MSVC2_NAME
      microservices:
      - name: $MSVC1_NAME
        rebuild: \"{{rebuild}}\"
        agent:
          name: \"{{$APP_TEMPLATE_KEY}}\"
          config:
            bluetoothEnabled: true # this will install the iofog/restblue microservice
            abstractedHardwareEnabled: false
        images:
          arm: edgeworx/healthcare-heart-rate:arm-v1
          x86: edgeworx/healthcare-heart-rate:x86-v1
          registry: remote # public docker
        container:
          rootHostAccess: \"{{root-host-access}}\"
          ports: []
        config:
          test_mode: true
          data_label: 'Anonymous_Person'
          first_custom: \"{{magic-number}}\"
          second_custom: \"{{turtle}}\"
      # Simple JSON viewer for the heart rate output
      - name: $MSVC2_NAME
        agent:
          name: \"{{$APP_TEMPLATE_KEY}}\"
        images:
          arm: edgeworx/healthcare-heart-rate-ui:arm
          x86: edgeworx/healthcare-heart-rate-ui:x86
          registry: remote
        container:
          rootHostAccess: false
          ports:
            # The ui will be listening on port 80 (internal).
            - external: \"{{external}}\"
              internal: \"{{internal}}\"
              public:
                schemes:
                - http
                protocol: http
                router:
                  port: \"{{public-port}}\"
          volumes:
          - hostDestination: $VOL_DEST
            containerDestination: $VOL_CONT_DEST
            accessMode: rw
          env:
            - key: BASE_URL
              value: http://localhost:8080/data" > test/conf/app-template.yaml
}
