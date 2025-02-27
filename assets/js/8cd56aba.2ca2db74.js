"use strict";(self.webpackChunkdocsite=self.webpackChunkdocsite||[]).push([[8053],{867:(e,r,n)=>{n.r(r),n.d(r,{assets:()=>l,contentTitle:()=>s,default:()=>p,frontMatter:()=>a,metadata:()=>i,toc:()=>c});var o=n(4848),t=n(8453);const a={sidebar_position:8},s="Cluster O&M operations",i={id:"manual/ob-operator-user-guide/cluster-management-of-ob-operator/cluster-operation",title:"Cluster O&M operations",description:"Introduction",source:"@site/docs/manual/500.ob-operator-user-guide/100.cluster-management-of-ob-operator/800.cluster-operation.md",sourceDirName:"manual/500.ob-operator-user-guide/100.cluster-management-of-ob-operator",slug:"/manual/ob-operator-user-guide/cluster-management-of-ob-operator/cluster-operation",permalink:"/ob-operator/docs/manual/ob-operator-user-guide/cluster-management-of-ob-operator/cluster-operation",draft:!1,unlisted:!1,editUrl:"https://github.com/oceanbase/ob-operator/tree/master/docsite/docs/manual/500.ob-operator-user-guide/100.cluster-management-of-ob-operator/800.cluster-operation.md",tags:[],version:"current",sidebarPosition:8,frontMatter:{sidebar_position:8},sidebar:"manualSidebar",previous:{title:"Delete a cluster",permalink:"/ob-operator/docs/manual/ob-operator-user-guide/cluster-management-of-ob-operator/delete-cluster"},next:{title:"Tenant management",permalink:"/ob-operator/docs/category/tenant-management"}},l={},c=[{value:"Introduction",id:"introduction",level:2},{value:"Operations",id:"operations",level:2},{value:"AddZones",id:"addzones",level:3},{value:"DeleteZones",id:"deletezones",level:3},{value:"AdjustReplicas",id:"adjustreplicas",level:3},{value:"Upgrade",id:"upgrade",level:3},{value:"RestartOBServers",id:"restartobservers",level:3},{value:"DeleteOBServers",id:"deleteobservers",level:3},{value:"ModifyOBServers",id:"modifyobservers",level:3},{value:"SetParameters",id:"setparameters",level:3}];function d(e){const r={admonition:"admonition",code:"code",h1:"h1",h2:"h2",h3:"h3",li:"li",p:"p",pre:"pre",ul:"ul",...(0,t.R)(),...e.components};return(0,o.jsxs)(o.Fragment,{children:[(0,o.jsx)(r.h1,{id:"cluster-om-operations",children:"Cluster O&M operations"}),"\n",(0,o.jsx)(r.h2,{id:"introduction",children:"Introduction"}),"\n",(0,o.jsx)(r.admonition,{type:"info",children:(0,o.jsx)(r.p,{children:"OBClusterOperation is a feature that is available in ob-operator v2.2.2 and later."})}),"\n",(0,o.jsxs)(r.p,{children:["In order to simplify the operation and maintenance of OceanBase clusters and keep the operation traces in short term, ob-operator provides the cluster O&M resource ",(0,o.jsx)(r.code,{children:"OBClusterOperation"})," for you to perform cluster O&M operations."]}),"\n",(0,o.jsxs)(r.p,{children:["By creating the ",(0,o.jsx)(r.code,{children:"OBClusterOperation"})," resource, you can perform the following cluster O&M operations:"]}),"\n",(0,o.jsxs)(r.ul,{children:["\n",(0,o.jsxs)(r.li,{children:[(0,o.jsx)(r.code,{children:"AddZones"}),": Add zones to the cluster."]}),"\n",(0,o.jsxs)(r.li,{children:[(0,o.jsx)(r.code,{children:"DeleteZones"}),": Delete zones from the cluster."]}),"\n",(0,o.jsxs)(r.li,{children:[(0,o.jsx)(r.code,{children:"AdjustReplicas"}),": Adjust the number of replicas of zones."]}),"\n",(0,o.jsxs)(r.li,{children:[(0,o.jsx)(r.code,{children:"Upgrade"}),": Upgrade the version of oceanbase cluster."]}),"\n",(0,o.jsxs)(r.li,{children:[(0,o.jsx)(r.code,{children:"RestartOBServers"}),": Restart the specific oceanbase servers."]}),"\n",(0,o.jsxs)(r.li,{children:[(0,o.jsx)(r.code,{children:"DeleteOBServers"}),": Delete the specific oceanbase servers."]}),"\n",(0,o.jsxs)(r.li,{children:[(0,o.jsx)(r.code,{children:"ModifyOBServers"}),": Modify the configuration of specific oceanbase servers, including cpu, memory, storage class, storage capacity, monitor deployment and NFS backup volume mount."]}),"\n",(0,o.jsxs)(r.li,{children:[(0,o.jsx)(r.code,{children:"SetParameters"}),": Set the parameters of oceanbase cluster."]}),"\n"]}),"\n",(0,o.jsxs)(r.p,{children:["The ",(0,o.jsx)(r.code,{children:"OBClusterOperation"})," resource is a custom resource with the following fields:"]}),"\n",(0,o.jsx)(r.pre,{children:(0,o.jsx)(r.code,{className:"language-yaml",children:"apiVersion: oceanbase.oceanbase.com/v1alpha1\nkind: OBClusterOperation\nmetadata:\n  generateName: <op-name>- # The name of the OBClusterOperation resource will be automatically generated by `kubectl create`.\n  namespace: <namespace>\nspec:\n  obcluster: <obcluster-name> # The name of the OBCluster resource to be operated.\n  type: <operation-type> # The type of the operation, including AddZones, DeleteZones, AdjustReplicas, Upgrade, RestartOBServers, DeleteOBServers, ModifyOBServers, SetParameters.\n  force: <force> # Whether to force the operation, default is false.\n  ttlDays: <ttlDays> # The number of days to keep the operation traces, default is 7.\n  <configuration-for-operation>: # The configuration for the operation, which is different for different operation types. The field name is the same as the operation type while the first capital letter is replaced with lowercase letter. For example, the configuration field for AddZones operation is addZones.\n    field1: value1\n    field2: value2\n    # ...\n"})}),"\n",(0,o.jsxs)(r.p,{children:["What needs to be noted is that only the specific configuration that matches the operation type will take effect. That is to say, if the operation type is ",(0,o.jsx)(r.code,{children:"AddZones"}),", only the ",(0,o.jsx)(r.code,{children:"addZones"})," field will take effect, and other specific configuration fields will be ignored."]}),"\n",(0,o.jsxs)(r.p,{children:["The ",(0,o.jsx)(r.code,{children:"OBClusterOperation"})," resource is a one-time resource, which means that the resource will be deleted automatically after the operation is completed. The operation traces will be kept for a period of time specified by the ",(0,o.jsx)(r.code,{children:"ttlDays"})," field."]}),"\n",(0,o.jsxs)(r.p,{children:["We recommend that you use the ",(0,o.jsx)(r.code,{children:"kubectl create"})," command to create the ",(0,o.jsx)(r.code,{children:"OBClusterOperation"})," resource to avoid applying resources with duplicated name, which can automatically generate the name of the resource with the ",(0,o.jsx)(r.code,{children:"generateName"})," field. For example,"]}),"\n",(0,o.jsx)(r.pre,{children:(0,o.jsx)(r.code,{className:"language-shell",children:"kubectl create -f path/to/obclusteroperation.yaml\n"})}),"\n",(0,o.jsx)(r.h2,{id:"operations",children:"Operations"}),"\n",(0,o.jsx)(r.h3,{id:"addzones",children:"AddZones"}),"\n",(0,o.jsx)(r.p,{children:"The configuration for the AddZones operation is as follows:"}),"\n",(0,o.jsx)(r.pre,{children:(0,o.jsx)(r.code,{className:"language-yaml",children:"apiVersion: oceanbase.oceanbase.com/v1alpha1\nkind: OBClusterOperation\nmetadata:\n  generateName: op-add-zones-\nspec:\n  obcluster: test\n  type: AddZones\n  addZones:\n    - zone: zone2\n      replica: 1\n    - zone: zone3\n      replica: 1\n"})}),"\n",(0,o.jsx)(r.h3,{id:"deletezones",children:"DeleteZones"}),"\n",(0,o.jsx)(r.p,{children:"The configuration for the DeleteZones operation is as follows:"}),"\n",(0,o.jsx)(r.pre,{children:(0,o.jsx)(r.code,{className:"language-yaml",children:"apiVersion: oceanbase.oceanbase.com/v1alpha1\nkind: OBClusterOperation\nmetadata:\n  generateName: op-delete-zones-\nspec:\n  obcluster: test\n  type: DeleteZones\n  deleteZones:\n    - zone2\n"})}),"\n",(0,o.jsx)(r.h3,{id:"adjustreplicas",children:"AdjustReplicas"}),"\n",(0,o.jsxs)(r.p,{children:["The configuration for the ",(0,o.jsx)(r.code,{children:"AdjustReplicas"})," operation is as follows:"]}),"\n",(0,o.jsx)(r.pre,{children:(0,o.jsx)(r.code,{className:"language-yaml",children:"apiVersion: oceanbase.oceanbase.com/v1alpha1\nkind: OBClusterOperation\nmetadata:\n  generateName: op-adjust-replicas-\nspec:\n  obcluster: test\n  type: AdjustReplicas\n  adjustReplicas:\n    - zones: [zone1]\n      to: 2\n"})}),"\n",(0,o.jsx)(r.h3,{id:"upgrade",children:"Upgrade"}),"\n",(0,o.jsxs)(r.p,{children:["The configuration for the ",(0,o.jsx)(r.code,{children:"Upgrade"})," operation is as follows:"]}),"\n",(0,o.jsx)(r.pre,{children:(0,o.jsx)(r.code,{className:"language-yaml",children:"apiVersion: oceanbase.oceanbase.com/v1alpha1\nkind: OBClusterOperation\nmetadata:\n  generateName: op-upgrade-\nspec:\n  obcluster: test\n  type: AdjustReplicas\n  upgrade:\n    image: xxx/xxxxx\n"})}),"\n",(0,o.jsx)(r.h3,{id:"restartobservers",children:"RestartOBServers"}),"\n",(0,o.jsxs)(r.p,{children:["The configuration for the ",(0,o.jsx)(r.code,{children:"RestartOBServers"})," operation is as follows:"]}),"\n",(0,o.jsx)(r.pre,{children:(0,o.jsx)(r.code,{className:"language-yaml",children:"apiVersion: oceanbase.oceanbase.com/v1alpha1\nkind: OBClusterOperation\nmetadata:\n  generateName: op-restart-observers-\nspec:\n  obcluster: test\n  type: RestartOBServers\n  restartOBServers:\n    observers: # The servers to be restarted, default is empty.\n      - observer-xxx-1\n      - observer-xxx-5\n    obzones: # The zones to which the servers belong, default is empty.\n      - zone1\n      - zone2\n    all: false # Whether to restart all the servers in the cluster, default is false.\n"})}),"\n",(0,o.jsx)(r.h3,{id:"deleteobservers",children:"DeleteOBServers"}),"\n",(0,o.jsxs)(r.p,{children:["The configuration for the ",(0,o.jsx)(r.code,{children:"DeleteOBServers"})," operation is as follows:"]}),"\n",(0,o.jsx)(r.pre,{children:(0,o.jsx)(r.code,{className:"language-yaml",children:"apiVersion: oceanbase.oceanbase.com/v1alpha1\nkind: OBClusterOperation\nmetadata:\n  generateName: op-delete-observers-\nspec:\n  obcluster: test\n  type: AdjustReplicas\n  observers:\n    - observer-xxx-1\n    - observer-xxx-5\n"})}),"\n",(0,o.jsx)(r.h3,{id:"modifyobservers",children:"ModifyOBServers"}),"\n",(0,o.jsx)(r.admonition,{type:"note",children:(0,o.jsx)(r.p,{children:"ModifyOBServers operation will rolling replace the servers in the cluster one by one. The operation will be completed after all the servers are replaced. The next observer will be replaced only after the previous observer is successfully replaced."})}),"\n",(0,o.jsxs)(r.p,{children:["The configuration for the ",(0,o.jsx)(r.code,{children:"ModifyOBServers"})," operation is as follows:"]}),"\n",(0,o.jsx)(r.pre,{children:(0,o.jsx)(r.code,{className:"language-yaml",children:"apiVersion: oceanbase.oceanbase.com/v1alpha1\nkind: OBClusterOperation\nmetadata:\n  generateName: op-modify-observers-\nspec:\n  obcluster: test\n  type: ModifyOBServers\n  modifyOBServers:\n    resource: # The resource configuration to be modified, default is empty.\n      cpu: 3\n      memory: 13Gi\n    expandStorageSize: # The storage capacity to be expanded, default is empty.\n      dataStorage: 100Gi\n      logStorage: 50Gi\n      redoLogStorage: 100Gi\n    modifyStorageClass: # The storage class to be modified, default is empty.\n      dataStorage: new-storage-class\n      logStorage: new-storage-class\n      redoLogStorage: new-storage-class\n    addingMonitor: # The monitor to be added, default is empty.\n      image: xxx/obagent:xxx\n      resource:\n        cpu: 1\n        memory: 1Gi\n    removeMonitor: true # Whether to remove the monitor, default is false.\n    addingBackupVolume: # The backup volume to be added, default is empty.\n      volume:\n        name: backup\n        nfs:\n          server: 1.2.3.4\n          path: /opt/nfs\n          readOnly: false\n    removeBackupVolume: true # Whether to remove the backup volume, default is false.\n"})}),"\n",(0,o.jsx)(r.h3,{id:"setparameters",children:"SetParameters"}),"\n",(0,o.jsxs)(r.p,{children:["The configuration for the ",(0,o.jsx)(r.code,{children:"SetParameters"})," operation is as follows:"]}),"\n",(0,o.jsx)(r.pre,{children:(0,o.jsx)(r.code,{className:"language-yaml",children:'apiVersion: oceanbase.oceanbase.com/v1alpha1\nkind: OBClusterOperation\nmetadata:\n  generateName: op-set-parameters-\nspec:\n  obcluster: test\n  type: SetParameters\n  setParameters: # The parameters to be set\n    - name: __min_full_resource_pool_memory\n      value: "3221225472"\n    - name: enable_syslog_recycle\n      value: "True"\n'})})]})}function p(e={}){const{wrapper:r}={...(0,t.R)(),...e.components};return r?(0,o.jsx)(r,{...e,children:(0,o.jsx)(d,{...e})}):d(e)}},8453:(e,r,n)=>{n.d(r,{R:()=>s,x:()=>i});var o=n(6540);const t={},a=o.createContext(t);function s(e){const r=o.useContext(a);return o.useMemo((function(){return"function"==typeof e?e(r):{...r,...e}}),[r,e])}function i(e){let r;return r=e.disableParentContext?"function"==typeof e.components?e.components(t):e.components||t:s(e.components),o.createElement(a.Provider,{value:r},e.children)}}}]);