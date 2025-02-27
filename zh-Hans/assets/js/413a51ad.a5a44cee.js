"use strict";(self.webpackChunkdocsite=self.webpackChunkdocsite||[]).push([[5907],{6498:(e,n,t)=>{t.r(n),t.d(n,{assets:()=>c,contentTitle:()=>i,default:()=>h,frontMatter:()=>s,metadata:()=>a,toc:()=>l});var r=t(4848),o=t(8453);const s={sidebar_position:2},i="\u521b\u5efa\u79df\u6237",a={id:"manual/ob-operator-user-guide/tenant-management-of-ob-operator/create-tenant",title:"\u521b\u5efa\u79df\u6237",description:"\u672c\u6587\u4ecb\u7ecd\u901a\u8fc7 ob-operator \u521b\u5efa\u79df\u6237\u3002",source:"@site/i18n/zh-Hans/docusaurus-plugin-content-docs/current/manual/500.ob-operator-user-guide/200.tenant-management-of-ob-operator/100.create-tenant.md",sourceDirName:"manual/500.ob-operator-user-guide/200.tenant-management-of-ob-operator",slug:"/manual/ob-operator-user-guide/tenant-management-of-ob-operator/create-tenant",permalink:"/ob-operator/zh-Hans/docs/manual/ob-operator-user-guide/tenant-management-of-ob-operator/create-tenant",draft:!1,unlisted:!1,editUrl:"https://github.com/oceanbase/ob-operator/tree/master/docsite/docs/manual/500.ob-operator-user-guide/200.tenant-management-of-ob-operator/100.create-tenant.md",tags:[],version:"current",sidebarPosition:2,frontMatter:{sidebar_position:2},sidebar:"manualSidebar",previous:{title:"\u79df\u6237\u7ba1\u7406",permalink:"/ob-operator/zh-Hans/docs/manual/ob-operator-user-guide/tenant-management-of-ob-operator/tenant-management-intro"},next:{title:"Modify tenant",permalink:"/ob-operator/zh-Hans/docs/category/modify-tenant"}},c={},l=[{value:"\u524d\u63d0\u6761\u4ef6",id:"\u524d\u63d0\u6761\u4ef6",level:2},{value:"\u4f7f\u7528\u914d\u7f6e\u6587\u4ef6\u521b\u5efa\u79df\u6237",id:"\u4f7f\u7528\u914d\u7f6e\u6587\u4ef6\u521b\u5efa\u79df\u6237",level:2},{value:"\u521b\u5efa\u79df\u6237\u793a\u4f8b",id:"\u521b\u5efa\u79df\u6237\u793a\u4f8b",level:2},{value:"\u786e\u8ba4\u79df\u6237\u662f\u5426\u521b\u5efa\u6210\u529f",id:"\u786e\u8ba4\u79df\u6237\u662f\u5426\u521b\u5efa\u6210\u529f",level:2},{value:"\u540e\u7eed\u64cd\u4f5c",id:"\u540e\u7eed\u64cd\u4f5c",level:2}];function d(e){const n={a:"a",code:"code",h1:"h1",h2:"h2",li:"li",p:"p",pre:"pre",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",ul:"ul",...(0,o.R)(),...e.components};return(0,r.jsxs)(r.Fragment,{children:[(0,r.jsx)(n.h1,{id:"\u521b\u5efa\u79df\u6237",children:"\u521b\u5efa\u79df\u6237"}),"\n",(0,r.jsx)(n.p,{children:"\u672c\u6587\u4ecb\u7ecd\u901a\u8fc7 ob-operator \u521b\u5efa\u79df\u6237\u3002"}),"\n",(0,r.jsx)(n.h2,{id:"\u524d\u63d0\u6761\u4ef6",children:"\u524d\u63d0\u6761\u4ef6"}),"\n",(0,r.jsx)(n.p,{children:"\u521b\u5efa\u79df\u6237\u524d\uff0c\u60a8\u9700\u8981\u786e\u4fdd\uff1a"}),"\n",(0,r.jsxs)(n.ul,{children:["\n",(0,r.jsxs)(n.li,{children:["\n",(0,r.jsx)(n.p,{children:"ob-operator v2.1.0 \u53ca\u4ee5\u4e0a\u3002"}),"\n"]}),"\n",(0,r.jsxs)(n.li,{children:["\n",(0,r.jsx)(n.p,{children:"OceanBase \u96c6\u7fa4\u90e8\u7f72\u5b8c\u6210\u4e14\u6b63\u5e38\u8fd0\u884c\u3002"}),"\n"]}),"\n"]}),"\n",(0,r.jsx)(n.h2,{id:"\u4f7f\u7528\u914d\u7f6e\u6587\u4ef6\u521b\u5efa\u79df\u6237",children:"\u4f7f\u7528\u914d\u7f6e\u6587\u4ef6\u521b\u5efa\u79df\u6237"}),"\n",(0,r.jsxs)(n.p,{children:["\u901a\u8fc7\u5e94\u7528\u79df\u6237\u914d\u7f6e\u6587\u4ef6\u521b\u5efa\u79df\u6237\u3002\u914d\u7f6e\u6587\u4ef6\u5185\u5bb9\u53ef\u53c2\u8003 ",(0,r.jsx)(n.a,{href:"https://github.com/oceanbase/ob-operator/blob/stable/example/tenant/tenant.yaml",children:"GitHub"})," \u3002"]}),"\n",(0,r.jsx)(n.p,{children:"\u521b\u5efa\u79df\u6237\u7684\u547d\u4ee4\u5982\u4e0b\uff0c\u8be5\u547d\u4ee4\u4f1a\u5728\u5f53\u524d Kubernetes \u96c6\u7fa4\u4e2d\u521b\u5efa\u4e00\u4e2a OBTenant \u79df\u6237\u7684\u8d44\u6e90\u3002"}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-shell",children:"kubectl apply -f tenant.yaml\n"})}),"\n",(0,r.jsx)(n.h2,{id:"\u521b\u5efa\u79df\u6237\u793a\u4f8b",children:"\u521b\u5efa\u79df\u6237\u793a\u4f8b"}),"\n",(0,r.jsx)(n.p,{children:"\u521b\u5efa\u540d\u4e3a t1 \u7684\u4e00\u4e2a 3 \u526f\u672c\u7684 MySQL \u79df\u6237\uff0c\u5e76\u6307\u5b9a\u5141\u8bb8\u4efb\u4f55\u5ba2\u6237\u7aef IP \u8fde\u63a5\u8be5\u79df\u6237\u3002"}),"\n",(0,r.jsxs)(n.p,{children:["\u521b\u5efa\u79df\u6237\u65f6\uff0cob-operator \u4f1a\u6839\u636e\u914d\u7f6e\u6587\u4ef6 ",(0,r.jsx)(n.code,{children:"tenant.yaml"})," \u4e2d\u7684 pools \u6309\u7167 zone \u6765\u521b\u5efa\u5bf9\u5e94\u7684 resource unit \u548c resource pool\u3002\u6839\u636e resource \u4e0b\u7684\u914d\u7f6e\u9879\u6765\u521b\u5efa resource unit \u5e76\u4ee5\u6b64\u4f5c\u4e3a\u8d44\u6e90\u89c4\u683c\u6765\u521b\u5efa resource pool\u3002"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-yaml",children:"apiVersion: oceanbase.oceanbase.com/v1alpha1  \nkind: OBTenant  \nmetadata:\n  name: t1\n  namespace: oceanbase\nspec: \n  obcluster: obcluster\n  tenantName: t1\n  unitNum: 1 \n  charset: utf8mb4  \n  connectWhiteList: '%'\n  forceDelete: true\n  credentials: # \u53ef\u9009\n    root: t1-ro # \u53ef\u9009\uff0c\u5982\u4e0d\u4f20\u5219 root \u7528\u6237\u5bc6\u7801\u4e3a\u7a7a\n    standbyRo: t1-ro # \u53ef\u9009\uff0c\u5982\u4e0d\u4f20\u5219\u81ea\u52a8\u521b\u5efa\n  pools:\n    - zone: zone1\n      type: \n        name: Full \n        replica: 1\n        isActive: true\n      resource:\n        maxCPU: 1\n        minCPU: 1\n        memorySize: 5Gi\n        maxIops: 1024\n        minIops: 1024\n        iopsWeight: 2\n        logDiskSize: 12Gi\n    - zone: zone2\n      type: \n        name: Full\n        replica: 1\n        isActive: true\n      resource:\n        maxCPU: 1 \n        minCPU: 1 \n        memorySize: 5Gi\n        maxIops: 1024\n        minIops: 1024\n        iopsWeight: 2\n        logDiskSize: 12Gi \n    - zone: zone3\n      type: \n        name: Full\n        replica: 1\n        isActive: true\n      priority: 3\n      resource:\n        maxCPU: 1 \n        minCPU: 1\n        memorySize: 5Gi\n        maxIops: 1024\n        minIops: 1024\n        iopsWeight: 2\n        logDiskSize: 12Gi \n"})}),"\n",(0,r.jsx)(n.p,{children:"\u914d\u7f6e\u9879\u8bf4\u660e\u5982\u4e0b\uff1a"}),"\n",(0,r.jsxs)(n.table,{children:[(0,r.jsx)(n.thead,{children:(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.th,{children:"\u914d\u7f6e\u9879"}),(0,r.jsx)(n.th,{children:"\u8bf4\u660e"})]})}),(0,r.jsxs)(n.tbody,{children:[(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"metadata.name"}),(0,r.jsx)(n.td,{children:"\u79df\u6237\u8d44\u6e90\u7684\u540d\u79f0\uff0c\u5728 K8s \u7684\u540c\u4e00\u4e2a\u547d\u540d\u7a7a\u95f4\u4e0b\u552f\u4e00\uff1b\u5fc5\u586b\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"metadata.namespace"}),(0,r.jsx)(n.td,{children:"\u6307\u5b9a\u79df\u6237\u8d44\u6e90\u6240\u5728\u7684\u547d\u540d\u7a7a\u95f4\uff1b\u5fc5\u586b\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"obcluster"}),(0,r.jsx)(n.td,{children:"\u6307\u5b9a\u9700\u8981\u521b\u5efa\u79df\u6237\u7684 OceanBase \u6570\u636e\u5e93\u96c6\u7fa4\u540d\uff1b\u5fc5\u586b\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"tenantName"}),(0,r.jsxs)(n.td,{children:["\u79df\u6237\u540d\u3002\u79df\u6237\u540d\u7684\u5408\u6cd5\u6027\u548c\u53d8\u91cf\u540d\u4e00\u81f4\uff0c\u6700\u957f 128 \u4e2a\u5b57\u7b26\uff0c\u5b57\u7b26\u53ea\u80fd\u662f\u5927\u5c0f\u5199\u82f1\u6587\u5b57\u6bcd\u3001\u6570\u5b57\u548c\u4e0b\u5212\u7ebf\uff0c\u800c\u4e14\u5fc5\u987b\u4ee5\u5b57\u6bcd\u6216\u4e0b\u5212\u7ebf\u5f00\u5934\uff0c\u5e76\u4e14\u4e0d\u80fd\u662f OceanBase \u6570\u636e\u5e93\u7684\u5173\u952e\u5b57\u3002 OceanBase \u6570\u636e\u5e93\u4e2d\u6240\u652f\u6301\u7684\u5173\u952e\u5b57\u8bf7\u53c2\u89c1 MySQL \u6a21\u5f0f\u7684 ",(0,r.jsx)(n.a,{href:"https://www.oceanbase.com/docs/common-oceanbase-database-cn-1000000000218216",children:"\u9884\u7559\u5173\u952e\u5b57"}),"\u548c Oracle \u6a21\u5f0f\u7684",(0,r.jsx)(n.a,{href:"https://www.oceanbase.com/docs/common-oceanbase-database-cn-1000000000218217",children:"\u9884\u7559\u5173\u952e\u5b57"}),"\uff1b\u5fc5\u586b\u3002"]})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"unitNum"}),(0,r.jsx)(n.td,{children:"\u6307\u5b9a\u8981\u521b\u5efa\u7684\u5355\u4e2a ZONE \u4e0b\u7684\u5355\u5143\u4e2a\u6570\uff0c\u53d6\u503c\u8981\u5c0f\u4e8e\u5355\u4e2a ZONE \u4e2d\u7684 OBServer \u8282\u70b9\u4e2a\u6570\uff1b\u5fc5\u586b\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"charset"}),(0,r.jsxs)(n.td,{children:["\u6307\u5b9a\u79df\u6237\u7684\u5b57\u7b26\u96c6\uff0c\u5b57\u7b26\u96c6\u76f8\u5173\u7684\u4ecb\u7ecd\u4fe1\u606f\u8bf7\u53c2\u89c1",(0,r.jsx)(n.a,{href:"https://www.oceanbase.com/docs/common-oceanbase-database-cn-1000000000221234",children:"\u5b57\u7b26\u96c6"}),"\uff1b\u975e\u5fc5\u586b\uff0c\u9ed8\u8ba4\u8bbe\u7f6e\u4e3a ",(0,r.jsx)(n.code,{children:"utf8mb4"}),"\u3002"]})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"collate"}),(0,r.jsxs)(n.td,{children:["\u6307\u5b9a\u79df\u6237\u7684\u5b57\u7b26\u5e8f\uff0c\u5b57\u7b26\u5e8f\u76f8\u5173\u7684\u4ecb\u7ecd\u4fe1\u606f\u8bf7\u53c2\u89c1",(0,r.jsx)(n.a,{href:"https://www.oceanbase.com/docs/common-oceanbase-database-cn-1000000000222182",children:"\u5b57\u7b26\u5e8f"}),"\uff1b\u975e\u5fc5\u586b\u3002"]})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"connectWhiteList"}),(0,r.jsxs)(n.td,{children:["\u6307\u5b9a\u5141\u8bb8\u8fde\u63a5\u8be5\u79df\u6237\u7684\u5ba2\u6237\u7aef IP\uff0c",(0,r.jsx)(n.code,{children:"%"})," \u8868\u793a\u4efb\u4f55\u5ba2\u6237\u7aef IP \u90fd\u53ef\u4ee5\u8fde\u63a5\u8be5\u79df\u6237\uff1b\u975e\u5fc5\u586b\uff0c\u9ed8\u8ba4\u8bbe\u7f6e\u4e3a ",(0,r.jsx)(n.code,{children:"%"}),"\u3002\u5982\u679c\u7528\u6237\u9700\u8981\u4fee\u6539\u6539\u914d\u7f6e\uff0c\u5219\u9700\u8981\u5c06 ob-operator \u6240\u5904\u7684\u7f51\u6bb5\u5305\u542b\u5728\u914d\u7f6e\u5185\uff0c\u5426\u5219 ob-operator \u4f1a\u8fde\u63a5\u4e0d\u4e0a\u8be5\u79df\u6237\u3002"]})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"forceDelete"}),(0,r.jsx)(n.td,{children:"\u5220\u9664\u65f6\u662f\u5426\u5f3a\u5236\u5220\u9664\uff0c\u975e\u5fc5\u586b\uff0c\u9ed8\u8ba4\u4e3a false\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"credentials"}),(0,r.jsx)(n.td,{children:"\u521b\u5efa\u79df\u6237\u65f6\u521b\u5efa\u7528\u6237\u548c\u4fee\u6539\u5bc6\u7801\u7684 Secret \u8d44\u6e90\u5f15\u7528\u3002\u76ee\u524d\u652f\u6301\u914d\u7f6e root \u8d26\u53f7\u548c standbyRo \u4e24\u4e2a\u7528\u6237\u7684\u5bc6\u7801\uff0c\u975e\u5fc5\u586b\uff0c\u4e0d\u586b\u5219\u4e0d\u4fee\u6539\u5bc6\u7801\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"pools"}),(0,r.jsx)(n.td,{children:"\u79df\u6237\u7684\u62d3\u6251\u7ed3\u6784\uff0c\u7528\u4e8e\u5b9a\u4e49\u79df\u6237\u5728\u6bcf\u4e2a zone \u4e0a\u7684\u526f\u672c\u3001\u8d44\u6e90\u5206\u5e03\u7b49\u60c5\u51b5\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"type.name"}),(0,r.jsx)(n.td,{children:"\u6307\u5b9a\u79df\u6237\u5728\u8be5 zone \u7684\u526f\u672c\u7c7b\u578b\uff0c\u652f\u6301 full \u548c readonly, \u9700\u8981\u5199\u51fa\u5b8c\u6574\u7c7b\u578b, \u5927\u5c0f\u5199\u4e0d\u654f\u611f\uff1b\u5fc5\u586b\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"type.replica"}),(0,r.jsx)(n.td,{children:"\u6307\u5b9a\u79df\u6237\u5728\u8be5 zone \u7684\u526f\u672c\u6570\uff1b\u975e\u5fc5\u586b\uff0c\u9ed8\u8ba4\u4e3a 1\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"type.isActive"}),(0,r.jsx)(n.td,{children:"\u662f\u5426\u542f\u7528 zone\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"priority"}),(0,r.jsx)(n.td,{children:"\u6307\u5b9a\u5f53\u524d zone \u7684\u4f18\u5148\u7ea7\uff0c\u6570\u5b57\u8d8a\u5927\u4f18\u5148\u7ea7\u8d8a\u9ad8\uff1b\u975e\u5fc5\u586b\uff0c\u9ed8\u8ba4\u4e3a 0\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"resource"}),(0,r.jsx)(n.td,{children:"\u6307\u5b9a\u79df\u6237\u5728\u8be5 zone \u7684\u8d44\u6e90\u60c5\u51b5\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"maxCPU"}),(0,r.jsx)(n.td,{children:"\u6307\u5b9a\u79df\u6237\u5728\u8be5 zone \u4e0a \u4f7f\u7528\u7684\u8d44\u6e90\u5355\u5143\u63d0\u4f9b\u7684 CPU \u7684\u4e0a\u9650\uff1b\u5fc5\u586b\uff0c\u6700\u5c0f\u503c\u4e3a 1\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"minCPU"}),(0,r.jsx)(n.td,{children:"\u6307\u5b9a\u79df\u6237\u5728\u8be5 zone \u4e0a \u4f7f\u7528\u7684\u8d44\u6e90\u5355\u5143\u63d0\u4f9b\u7684 CPU \u7684\u4e0b\u9650\uff1b\u975e\u5fc5\u586b\uff0c\u9ed8\u8ba4\u7b49\u4e8e maxCPU\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"memorySize"}),(0,r.jsx)(n.td,{children:"\u6307\u5b9a\u79df\u6237\u5728\u8be5 zone \u4e0a \u4f7f\u7528\u7684\u8d44\u6e90\u5355\u5143\u63d0\u4f9b\u7684 Memory \u7684\u5927\u5c0f\uff1b\u5fc5\u586b\uff0c\u6700\u5c0f\u503c\u4e3a 1Gi\uff1b\u6ce8\u610f\u96c6\u7fa4\u7684 __min_full_resource_pool_memory \u914d\u7f6e\u9879\u7684\u503c"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"maxIops"}),(0,r.jsx)(n.td,{children:"\u6307\u5b9a\u79df\u6237\u5728\u8be5 zone \u4e0a \u4f7f\u7528\u7684\u8d44\u6e90\u5355\u5143\u63d0\u4f9b\u7684 Iops \u7684\u4e0a\u9650\uff1b\u975e\u5fc5\u586b\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"minIops"}),(0,r.jsx)(n.td,{children:"\u6307\u5b9a\u79df\u6237\u5728\u8be5 zone \u4e0a \u4f7f\u7528\u7684\u8d44\u6e90\u5355\u5143\u63d0\u4f9b\u7684 Iops \u7684\u4e0b\u9650\uff1b\u975e\u5fc5\u586b\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"iopsWeight"}),(0,r.jsx)(n.td,{children:"\u6307\u5b9a\u79df\u6237\u5728\u8be5 zone \u4e0a \u4f7f\u7528\u7684\u8d44\u6e90\u5355\u5143\u63d0\u4f9b\u7684 Iops \u6743\u91cd\u3002\u975e\u5fc5\u586b\uff0c\u9ed8\u8ba4\u7b49\u4e8e 1\u3002"})]}),(0,r.jsxs)(n.tr,{children:[(0,r.jsx)(n.td,{children:"logDiskSize"}),(0,r.jsx)(n.td,{children:"\u6307\u5b9a\u79df\u6237\u5728\u8be5 zone \u4e0a \u4f7f\u7528\u7684\u8d44\u6e90\u5355\u5143\u63d0\u4f9b\u7684\u65e5\u5fd7\u76d8\u89c4\u683c\u3002\u975e\u5fc5\u586b\uff0c\u9ed8\u8ba4\u7b49\u4e8e 3 \u500d\u7684\u5185\u5b58\u89c4\u683c\uff0c\u6700\u5c0f\u503c\u4e3a 2Gi\u3002"})]})]})]}),"\n",(0,r.jsx)(n.h2,{id:"\u786e\u8ba4\u79df\u6237\u662f\u5426\u521b\u5efa\u6210\u529f",children:"\u786e\u8ba4\u79df\u6237\u662f\u5426\u521b\u5efa\u6210\u529f"}),"\n",(0,r.jsxs)(n.p,{children:["\u521b\u5efa\u79df\u6237\u540e\uff0c\u6267\u884c\u4ee5\u4e0b\u8bed\u53e5\uff0c\u67e5\u770b\u5f53\u524d Kubernetes \u96c6\u7fa4\u4e2d\u662f\u5426\u6709\u65b0\u521b\u5efa\u7684\u79df\u6237\u7684 OBTenant \u8d44\u6e90\uff0c\u5e76\u4e14\u8be5 OBTenant \u8d44\u6e90\u7684 ",(0,r.jsx)(n.code,{children:"Status.status"})," \u4e3a ",(0,r.jsx)(n.code,{children:"running"}),"\uff0c\u76f8\u5173\u914d\u7f6e\u90fd\u4f1a\u5728 Status \u4e2d\u5c55\u793a\u3002"]}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-shell",children:"kubectl describe obtenants.oceanbase.oceanbase.com -n oceanbase t1\n"})}),"\n",(0,r.jsx)(n.p,{children:"\u8fd4\u56de\u7684\u793a\u4f8b\u7ed3\u679c\u5982\u4e0b\uff1a"}),"\n",(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-shell",children:"Name:         t1\nNamespace:    oceanbase\nLabels:       <none>\nAnnotations:  <none>\nAPI Version:  oceanbase.oceanbase.com/v1alpha1\nKind:         OBTenant\nMetadata:\n  Creation Timestamp:  2023-11-13T07:28:31Z\n  Finalizers:\n    finalizers.oceanbase.com.deleteobtenant\n  Generation:        2\n  Resource Version:  940236\n  UID:               34036a49-26bf-47cf-8201-444b3850aaa2\nSpec:\n  Charset:             utf8mb4\n  Connect White List:  %\n  Credentials:\n    Root:        t1-ro\n    Standby Ro:  t1-ro\n  Force Delete:  true\n  Obcluster:     obcluster\n  Pools:\n    Priority:  1\n    Resource:\n      Iops Weight:    2\n      Log Disk Size:  12Gi\n      Max CPU:        1\n      Max Iops:       1024\n      Memory Size:    5Gi\n      Min CPU:        1\n      Min Iops:       1024\n    Type:\n      Is Active:  true\n      Name:       Full\n      Replica:    1\n    Zone:         zone1\n    Priority:     1\n    Resource:\n      Iops Weight:    2\n      Log Disk Size:  12Gi\n      Max CPU:        1\n      Max Iops:       1024\n      Memory Size:    5Gi\n      Min CPU:        1\n      Min Iops:       1024\n    Type:\n      Is Active:  true\n      Name:       Full\n      Replica:    1\n    Zone:         zone2\n    Priority:     3\n    Resource:\n      Iops Weight:    2\n      Log Disk Size:  12Gi\n      Max CPU:        1\n      Max Iops:       1024\n      Memory Size:    5Gi\n      Min CPU:        1\n      Min Iops:       1024\n    Type:\n      Is Active:  true\n      Name:       Full\n      Replica:    1\n    Zone:         zone3\n  Tenant Name:    t1\n  Tenant Role:    PRIMARY\n  Unit Num:       1\nStatus:\n  Credentials:\n    Root:        t1-ro\n    Standby Ro:  t1-ro\n  Resource Pool:\n    Priority:  1\n    Type:\n      Is Active:  true\n      Name:       FULL\n      Replica:    1\n    Unit Config:\n      Iops Weight:    2\n      Log Disk Size:  12884901888\n      Max CPU:        1\n      Max Iops:       1024\n      Memory Size:    5368709120\n      Min CPU:        1\n      Min Iops:       1024\n    Unit Num:         1\n    Units:\n      Migrate:\n        Server IP:    \n        Server Port:  0\n      Server IP:      10.42.0.189\n      Server Port:    2882\n      Status:         ACTIVE\n      Unit Id:        1006\n    Zone List:        zone1\n    Priority:         1\n    Type:\n      Is Active:  true\n      Name:       FULL\n      Replica:    1\n    Unit Config:\n      Iops Weight:    2\n      Log Disk Size:  12884901888\n      Max CPU:        1\n      Max Iops:       1024\n      Memory Size:    5368709120\n      Min CPU:        1\n      Min Iops:       1024\n    Unit Num:         1\n    Units:\n      Migrate:\n        Server IP:    \n        Server Port:  0\n      Server IP:      10.42.1.118\n      Server Port:    2882\n      Status:         ACTIVE\n      Unit Id:        1007\n    Zone List:        zone2\n    Priority:         2\n    Type:\n      Is Active:  true\n      Name:       FULL\n      Replica:    1\n    Unit Config:\n      Iops Weight:    2\n      Log Disk Size:  12884901888\n      Max CPU:        1\n      Max Iops:       1024\n      Memory Size:    5368709120\n      Min CPU:        1\n      Min Iops:       1024\n    Unit Num:         1\n    Units:\n      Migrate:\n        Server IP:    \n        Server Port:  0\n      Server IP:      10.42.0.190\n      Server Port:    2882\n      Status:         ACTIVE\n      Unit Id:        1008\n    Zone List:        zone3\n  Status:             running\n  Tenant Record Info:\n    Charset:             utf8mb4\n    Connect White List:  %\n    Locality:            FULL{1}@zone1, FULL{1}@zone2, FULL{1}@zone3\n    Pool List:           pool_t1_zone1,pool_t1_zone2,pool_t1_zone3\n    Primary Zone:        zone3;zone1,zone2\n    Tenant ID:           1006\n    Unit Num:            1\n    Zone List:           zone1,zone2,zone3\n  Tenant Role:           PRIMARY\nEvents:\n  Type    Reason  Age                    From                 Message\n  ----    ------  ----                   ----                 -------\n  Normal          2m58s                  obtenant-controller  start creating\n  Normal          115s                   obtenant-controller  create OBTenant successfully\n"})}),"\n",(0,r.jsx)(n.h2,{id:"\u540e\u7eed\u64cd\u4f5c",children:"\u540e\u7eed\u64cd\u4f5c"}),"\n",(0,r.jsxs)(n.p,{children:["\u79df\u6237\u521b\u5efa\u6210\u529f\u540e\uff0c\u5176\u7ba1\u7406\u5458\u8d26\u53f7\u5bc6\u7801\u4e3a ",(0,r.jsx)(n.code,{children:"spec.credentials.root"})," \u5b57\u6bb5\u6307\u5b9a\u7684 secret \u4e2d\u5305\u542b\u7684\u5185\u5bb9\uff0c\u82e5\u521b\u5efa\u65f6\u6ca1\u6709\u6307\u5b9a\uff0c\u5219\u5bc6\u7801\u4e3a\u7a7a\u3002\u60a8\u53ef\u4ee5\u4f7f\u7528 ",(0,r.jsx)(n.code,{children:"obclient -h${podIP} -P2881 -uroot@tenantname -p -A"})," \u6216\u8005 ",(0,r.jsx)(n.code,{children:"mysql -h${podIP} -P2881 -uroot@tenantname -p -A"})," \u8bed\u53e5\u767b\u5f55\u6570\u636e\u5e93\u3002"]})]})}function h(e={}){const{wrapper:n}={...(0,o.R)(),...e.components};return n?(0,r.jsx)(n,{...e,children:(0,r.jsx)(d,{...e})}):d(e)}},8453:(e,n,t)=>{t.d(n,{R:()=>i,x:()=>a});var r=t(6540);const o={},s=r.createContext(o);function i(e){const n=r.useContext(s);return r.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function a(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(o):e.components||o:i(e.components),r.createElement(s.Provider,{value:n},e.children)}}}]);