    sed -i "s|AGENT_LIST=.*|AGENT_LIST=\"${{ env.agent_vm_list }}\"|g" test/env.sh
    sed -i "s|VANILLA_CONTROLLER=.*|VANILLA_CONTROLLER=\"${{ env.controller_vm }}\"|g" test/env.sh
    sed -i "s|NAMESPACE=.*|NAMESPACE=\"${{ github.run_number }}\"|g" test/env.sh
    sed -i "s|CONTROLLER_IMAGE=.*|CONTROLLER_IMAGE=\"${{ env.controller_image }}\"|g" test/env.sh
    sed -i "s|CONTROLLER_VANILLA_VERSION=.*|CONTROLLER_VANILLA_VERSION=\"${{ env.controller_version }}\"|g" test/env.sh
    sed -i "s|OPERATOR_IMAGE=.*|OPERATOR_IMAGE=\"${{ env.operator_image }}\"|g" test/env.sh
    sed -i "s|KUBELET_IMAGE=.*|KUBELET_IMAGE=\"${{ env.kubelet_image }}\"|g" test/env.sh
    sed -i "s|PORT_MANAGER_IMAGE=.*|PORT_MANAGER_IMAGE=\"${{ env.port_manager_image }}\"|g" test/env.sh
    sed -i "s|AGENT_IMAGE=.*|AGENT_IMAGE=\"${{ env.agent_image }}\"|g" test/env.sh
    sed -i "s|ROUTER_IMAGE=.*|ROUTER_IMAGE=\"${{ env.router_image }}\"|g" test/env.sh
    sed -i "s|ROUTER_ARM_IMAGE=.*|ROUTER_ARM_IMAGE=\"${{ env.router_arm_image }}\"|g" test/env.sh
    sed -i "s|PROXY_IMAGE=.*|PROXY_IMAGE=\"${{ env.proxy_image }}\"|g" test/env.sh
    sed -i "s|PROXY_ARM_IMAGE=.*|PROXY_ARM_IMAGE=\"${{ env.proxy_arm_image }}\"|g" test/env.sh
    sed -i "s|AGENT_VANILLA_VERSION=.*|AGENT_VANILLA_VERSION=\"${{ env.iofog_agent_version }}\"|g" test/env.sh
    sed -i "s|CONTROLLER_PACKAGE_CLOUD_TOKEN=.*|CONTROLLER_PACKAGE_CLOUD_TOKEN=\"${{ env.pkg.controller.token }}\"|g" test/env.sh
    sed -i "s|AGENT_PACKAGE_CLOUD_TOKEN=.*|AGENT_PACKAGE_CLOUD_TOKEN=\"${{ env.pkg.agent.token }}\"|g" test/env.sh
    cp test/env.sh test/conf
    cat test/conf/env.sh