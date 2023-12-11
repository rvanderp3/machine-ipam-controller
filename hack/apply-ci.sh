
echo "retrieving IPAM controller CI configuration"

curl https://raw.githubusercontent.com/rvanderp3/machine-ipam-controller/main/hack/ci-resources.yaml | envsubst | oc create -f -

echo "applying ippool configuration to compute machineset"
oc get machineset.machine.openshift.io -n openshift-machine-api -o json | jq -r '.items[0].spec.template.spec.providerSpec.value.network.devices[0] += 
{
    addressesFromPool: 
        [
            {
                group: "ipamcontroller.openshift.io", 
                name: "static-ci-pool", 
                resource: "IPPool"
            }
        ],
    nameservers:
        [ "8.8.8.8" ]
}' | jq '.items[1].spec.template.metadata.labels += 
{
    ipam: "true"
}' | oc apply -f -

echo "scaling up machineset with ippool configuration"
MACHINESET_NAME=$(oc get machineset.machine.openshift.io -n openshift-machine-api -o json | jq -r '.items[0].metadata.name')
oc scale machineset.machine.openshift.io --replicas=2 ${MACHINESET_NAME} -n openshift-machine-api

VALID_STATIC_IP=("192.168.${third_octet}.129" "192.168.${third_octet}.130" "192.168.${third_octet}.131")

echo "validating static IPs are applied to applicable nodes"
for retries in {1..16}; do
    NODEREFS=($(oc get machines.machine.openshift.io -n openshift-machine-api -l ipam=true -o=json | jq -r .items[].status.nodeRef.name))
    if [[ ${#NODEREFS[@]} -lt 2 ]]; then
        echo "${#NODEREFS[@]} of 2 node refs available"    
    else    
        NODES_VALIDATED=0
        for NODE in ${NODEREFS[@]}; do           
            if [[ ${NODE} = "null" ]]; then                                
                echo "not all machines have nodeRefs. Will recheck in 15 seconds."
                NODES_VALIDATED=$((${NODES_VALIDATED}-1))
                break
            fi
            echo "verifying static IP for node ${NODE}"        
            ADDRESS=$(oc get node ${NODE} -o=jsonpath='{.status.addresses}' | jq -r '.[] | select(.type=="InternalIP") | .address')
            if [ -z "${ADDRESS}" ]; then
                echo "no address available for node ${NODE}"
                break
            fi
            MATCH=0
            for VALID_IP in ${VALID_STATIC_IP[@]}; do
                if [[ ${VALID_IP} = ${ADDRESS} ]]; then
                    MATCH=1 
                fi
            done
            if [[ ${MATCH} -eq 0 ]]; then
                echo "node ${NODE} does not have an expected address. InternalIP ${ADDRESS}"
                NODES_VALIDATED=$((${NODES_VALIDATED}-1))
            fi        
        done
        if [[ ${NODES_VALIDATED} -eq 0 ]]; then
            echo "all nodes validated"
            exit 0
        else
            echo "not all nodes have been validated. Will recheck in 15 seconds"
            sleep 15
        fi
    fi
done 

echo "unable to verify applicable nodes received static IPs"
exit 1 