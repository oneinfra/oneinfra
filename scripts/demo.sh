bail_out() {
    echo "${1}"
    exit 1
}

if ! which docker &> /dev/null; then
    bail_out "Please, install docker in order to continue"
fi

if ! which kind &> /dev/null;  then
    bail_out "Please, install kind in order to continue"
fi

if ! which kubectl &> /dev/null; then
    bail_out "Please, install kubectl in order to continue"
fi

if ! which oi-local-hypervisor-set &> /dev/null; then
    bail_out "Please, install oi-local-hypervisor-set in order to continue"
fi

execute_command() {
    echo "** ${1}"
    if ! eval "${2}"; then
        bail_out "Failed to execute command ${2}"
    fi
}

wait_for_ns() {
    execute_command "Waiting for all deployments in namespace ${1} to be ready..." "kubectl wait --for=condition=Available deployment --timeout=2m -n ${1} --all"
}

execute_command "Creating kind cluster..." "kind create cluster"
execute_command "Deploying cert-manager..." "kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v0.14.1/cert-manager.yaml"
wait_for_ns cert-manager
execute_command "Deploying oneinfra 20.05.0-alpha13..." "kubectl apply -f https://raw.githubusercontent.com/oneinfra/oneinfra/20.05.0-alpha13/config/generated/all.yaml"
wait_for_ns oneinfra-system
execute_command "Creating fake local hypervisors..." "oi-local-hypervisor-set create --tcp | kubectl apply -f -"
execute_command "Generating oneinfra console JWT key..." "kubectl create secret generic -n oneinfra-system jwt-key --from-literal=jwt-key=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 64 | head -n 1)"
execute_command "Deploying oneinfra console 20.05.0-alpha3..." "kubectl apply -f https://raw.githubusercontent.com/oneinfra/console/20.05.0-alpha3/config/generated/all-kubernetes-secrets.yaml"
wait_for_ns oneinfra-system
echo "======================================="
echo "==    Visit http://localhost:8000    =="
echo "==    Username: sample-user          =="
echo "==    Password: sample-user          =="
echo "======================================="
kubectl port-forward -n oneinfra-system svc/oneinfra-console 8000:80
