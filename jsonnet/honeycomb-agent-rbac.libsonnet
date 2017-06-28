// CHANGE THIS IMPORT TO POINT TO YOUR LOCAL KSONNET
local k = import "/Users/jyao/heptio/hausdorff-ksonnet/ksonnet.beta.2/k.libsonnet";

// Destructuring imports.
local svcAccount = k.core.v1.serviceAccount;
local clRoleBinding = k.rbac.v1beta1.clusterRoleBinding;
local clRole = k.rbac.v1beta1.clusterRole;
local subject = clRoleBinding.subjectsType;
local rule = clRole.rulesType;

{
    getRbacComponents(name, namespace)::
        local metadata = svcAccount.mixin.metadata.name(name) +
            svcAccount.mixin.metadata.namespace(namespace);

        local hcServiceAccount = svcAccount.new() +
             metadata;
        
        local hcClusterRole = clRole.new() +
            metadata +
            clRole.rules(
                rule.new() +
                rule.apiGroups("*") +
                rule.resources("pods") +
                rule.verbs(["list", "watch"])
            );
        
        local hcClusterRoleBinding = clRoleBinding.new() + 
            clRoleBinding.mixin.metadata.name(name) +
            clRoleBinding.mixin.roleRef.apiGroup("rbac.authorization.k8s.io") + 
            clRoleBinding.mixin.roleRef.name(name) +
            clRoleBinding.subjects(
                subject.new() +
                subject.name(name) + 
                subject.namespace(namespace)
            );

        k.core.v1.list.new([hcServiceAccount, hcClusterRole, hcClusterRoleBinding]),
}
