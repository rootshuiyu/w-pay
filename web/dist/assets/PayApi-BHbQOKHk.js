import{_ as N,q as B,o as _,c as x,a as s,w as t,g as y,r,m as F,b as a,F as J,x as h,i as T,e as l,k as $,t as p,E as j}from"./index-Cik_B3pA.js";import{l as W,b as G,d as M}from"./platforms-BLct4fCF.js";import"./api-CPwamEKB.js";const Q={class:"api-page"},L={class:"doc-toolbar"},U={key:0,class:"doc-meta"},Y={class:"code"},Z={class:"code"},ee={class:"code"},ae=`{
  "amount": 100,
  "pay_type": 1,
  "pay_scene": "h5",
  "return_url": "https://cashier.example.com/done",
  "subject": "堂食消费"
}
// Header: X-App-Key: pk_xxxxxxxx（由我方分配）`,le=`{
  "code": 200,
  "message": "success",
  "data": {
    "order_id": "2082511055136755712",
    "pay_scene": "h5",
    "pay_url": "https://wx.tenpay.com/...",
    "qr_code_url": "",
    "amount": 100,
    "pay_type": 1
  }
}`,te=`{
  "code": 400,
  "message": "支付通道暂不可用，请稍后重试"
}`,se=`{
  "code": 200,
  "message": "success",
  "data": {
    "order_id": "2082511055136755712",
    "order_status": 0,
    "total_amount": 100,
    "pay_amount": 0,
    "pay_type": 1,
    "pay_scene": "h5",
    "qr_code_url": "...",
    "pay_time": null,
    "transaction_id": ""
  }
}`,oe={__name:"PayApi",setup(ne){const m=window.location.origin,g=y([]),v=y(""),n=F(()=>g.value.find(o=>String(o.id)===String(v.value)));function P(){if(n.value){const o=n.value.app_key;q.value=A(o),w.value=C(o),k.value=K(o)}}function A(o){return`curl -X POST ${m}/api/pay/create \\
  -H "Content-Type: application/json" \\
  -H "X-App-Key: ${o||"pk_your_app_key"}" \\
  -d '{"amount":100,"pay_type":1,"pay_scene":"h5","return_url":"https://cashier.example.com/done","subject":"test"}'`}function C(o){return`curl "${m}/api/pay/query?order_no=ORDER_ID" \\
  -H "X-App-Key: ${o||"pk_your_app_key"}"
# 或 order_id=ORDER_ID`}function K(o){return`const base = '${m}';
const appKey = '${o||"pk_your_app_key"}';
const r = await fetch(base + '/api/pay/create', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-App-Key': appKey,
  },
  body: JSON.stringify({
    amount: 100,           // 分
    pay_type: 1,           // 1微信 2支付宝
    pay_scene: 'h5',
    return_url: location.href,
    subject: '收款',
  }),
}).then(res => res.json());
if (r.code !== 200) throw new Error(r.message);
const d = r.data;
if (d.pay_url) location.href = d.pay_url;
else if (d.qr_code_url) showQRCode(d.qr_code_url);

// 查单确认（同样携带 X-App-Key）
const timer = setInterval(async () => {
  const q = await fetch(base + '/api/pay/query?order_no=' + d.order_id, {
    headers: { 'X-App-Key': appKey },
  }).then(x => x.json());
  if (q.data?.order_status === 1) { clearInterval(timer); onPaid(); }
}, 2000);`}const q=y(A("pk_your_app_key")),w=y(C("pk_your_app_key")),k=y(K("pk_your_app_key")),I=[{name:"amount",req:"是",desc:"金额（分），须 > 0"},{name:"pay_type",req:"是",desc:"1=微信，2=支付宝"},{name:"pay_scene",req:"否",desc:"h5（默认）| native；H5 失败自动降级 native"},{name:"return_url",req:"建议",desc:"H5/WAP 支付完成回跳"},{name:"quit_url",req:"否",desc:"支付宝取消支付回跳"},{name:"subject",req:"否",desc:"备注；也可用 biz_remark"},{name:"device_sn",req:"否",desc:"收银设备号"},{name:"store_id",req:"否",desc:"可选业务门店标识"},{name:"channel_type",req:"否",desc:"兼容：wechat / alipay"}],D=[{name:"order_id",desc:"订单号（create/query 通用）"},{name:"order_status",desc:"query：0待支付 1已支付 2已关闭 3退款"},{name:"total_amount",desc:"query：下单金额（分）"},{name:"pay_amount",desc:"query：实付金额（分），支付后更新"},{name:"pay_scene",desc:"实际支付方式 h5|native"},{name:"pay_url",desc:"create：H5/WAP 跳转链接"},{name:"qr_code_url",desc:"create/query：扫码链接（native 或兜底）"},{name:"amount",desc:"create：下单金额（分）"},{name:"pay_type",desc:"1=微信 2=支付宝"},{name:"pay_time",desc:"query：支付完成时间"},{name:"transaction_id",desc:"query：第三方支付流水号"}];function f(o){var b;const e=typeof o=="string"?o:(o==null?void 0:o.value)??String(o);(b=navigator.clipboard)==null||b.writeText(e),j.success("已复制")}function O(){if(!n.value)return;const o=G(n.value,m),e=(n.value.platform_name||"platform").replace(/[^\w\u4e00-\u9fa5-]+/g,"_");M(`${e}-收银API对接说明.md`,o),j.success("文档已下载")}return B(async()=>{const{list:o}=await W(1);g.value=o,o.length&&(v.value=o[0].id,P())}),(o,e)=>{const b=r("el-option"),R=r("el-select"),X=r("Download"),H=r("el-icon"),i=r("el-button"),u=r("el-card"),z=r("el-alert"),E=r("el-col"),c=r("el-table-column"),S=r("el-table"),V=r("el-row");return _(),x("div",Q,[s(u,{shadow:"never",class:"page-card",style:{"margin-bottom":"16px"}},{default:t(()=>[a("div",L,[s(R,{modelValue:v.value,"onUpdate:modelValue":e[0]||(e[0]=d=>v.value=d),filterable:"",placeholder:"选择代收平台（导出专属文档）",style:{width:"280px"},onChange:P},{default:t(()=>[(_(!0),x(J,null,h(g.value,d=>(_(),T(b,{key:d.id,label:d.platform_name,value:d.id},null,8,["label","value"]))),128))]),_:1},8,["modelValue"]),s(i,{type:"primary",disabled:!n.value,onClick:O},{default:t(()=>[s(H,null,{default:t(()=>[s(X)]),_:1}),e[5]||(e[5]=l("导出对接文档 (.md) ",-1))]),_:1},8,["disabled"]),n.value?(_(),T(i,{key:0,onClick:e[1]||(e[1]=d=>f(n.value.app_key))},{default:t(()=>[...e[6]||(e[6]=[l("复制 App Key",-1)])]),_:1})):$("",!0)]),n.value?(_(),x("p",U,[e[7]||(e[7]=l(" 当前：",-1)),a("strong",null,p(n.value.platform_name),1),e[8]||(e[8]=l(" · App Key ",-1)),a("code",null,p(n.value.app_key),1),e[9]||(e[9]=l(" · IP ",-1)),a("code",null,p(n.value.allowed_ips||"未配置"),1)])):$("",!0)]),_:1}),s(z,{type:"success",closable:!1,"show-icon":"",title:"标准聚合支付接口：创建订单 → 跳转 pay_url 或展示 qr_code_url → 查单确认。须携带 X-App-Key，且来源 IP 须在该平台对接白名单内。手机浏览器推荐 pay_scene=h5；PC/收款码屏用 native。",style:{"margin-bottom":"16px"}}),s(V,{gutter:16},{default:t(()=>[s(E,{xs:24,lg:14},{default:t(()=>[s(u,{shadow:"never",class:"page-card"},{header:t(()=>[...e[10]||(e[10]=[a("span",{class:"endpoint-title"},"POST /api/pay/create",-1),l(" — 创建支付订单",-1)])]),default:t(()=>[e[12]||(e[12]=a("p",{class:"desc"},[l(" 传入金额与支付方式即可出单。系统返回 "),a("code",null,"pay_url"),l("（H5/WAP 跳转）或 "),a("code",null,"qr_code_url"),l("（扫码），对接方按 "),a("code",null,"pay_scene"),l(" 处理即可。 ")],-1)),e[13]||(e[13]=a("div",{class:"label"},"请求体 JSON",-1)),a("pre",{class:"code"},p(ae)),e[14]||(e[14]=a("div",{class:"label"},"成功响应（外层统一格式）",-1)),a("pre",{class:"code"},p(le)),e[15]||(e[15]=a("div",{class:"label"},"失败响应示例",-1)),a("pre",{class:"code"},p(te)),e[16]||(e[16]=a("div",{class:"label"},"curl",-1)),a("pre",Y,p(q.value),1),s(i,{size:"small",onClick:e[2]||(e[2]=d=>f(q.value))},{default:t(()=>[...e[11]||(e[11]=[l("复制 curl",-1)])]),_:1})]),_:1}),s(u,{shadow:"never",class:"page-card",style:{"margin-top":"16px"}},{header:t(()=>[...e[17]||(e[17]=[a("span",{class:"endpoint-title"},"GET /api/pay/query",-1),l(" — 查询订单",-1)])]),default:t(()=>[e[19]||(e[19]=a("p",{class:"desc"},[l("参数 "),a("code",null,"order_no"),l(" 或 "),a("code",null,"order_id"),l("（二选一）。支付完成后建议定时查单，直至 "),a("code",null,"order_status=1"),l("。")],-1)),a("pre",Z,p(w.value),1),e[20]||(e[20]=a("div",{class:"label"},"成功响应 data",-1)),a("pre",{class:"code"},p(se)),s(i,{size:"small",onClick:e[3]||(e[3]=d=>f(w.value))},{default:t(()=>[...e[18]||(e[18]=[l("复制 curl",-1)])]),_:1})]),_:1}),s(u,{shadow:"never",class:"page-card",style:{"margin-top":"16px"}},{header:t(()=>[...e[21]||(e[21]=[l("收银端接入示例（JavaScript）",-1)])]),default:t(()=>[a("pre",ee,p(k.value),1),s(i,{size:"small",onClick:e[4]||(e[4]=d=>f(k.value))},{default:t(()=>[...e[22]||(e[22]=[l("复制代码",-1)])]),_:1})]),_:1})]),_:1}),s(E,{xs:24,lg:10},{default:t(()=>[s(u,{shadow:"never",class:"page-card"},{header:t(()=>[...e[23]||(e[23]=[l("create 请求字段",-1)])]),default:t(()=>[s(S,{data:I,size:"small",stripe:""},{default:t(()=>[s(c,{prop:"name",label:"字段",width:"110"}),s(c,{prop:"req",label:"必填",width:"52"}),s(c,{prop:"desc",label:"说明"})]),_:1})]),_:1}),s(u,{shadow:"never",class:"page-card",style:{"margin-top":"16px"}},{header:t(()=>[...e[24]||(e[24]=[l("create / query 响应字段",-1)])]),default:t(()=>[s(S,{data:D,size:"small",stripe:""},{default:t(()=>[s(c,{prop:"name",label:"字段",width:"120"}),s(c,{prop:"desc",label:"说明"})]),_:1})]),_:1}),s(u,{shadow:"never",class:"page-card",style:{"margin-top":"16px"}},{header:t(()=>[...e[25]||(e[25]=[l("流程说明",-1)])]),default:t(()=>[e[26]||(e[26]=a("ul",{class:"flow"},[a("li",null,[l("POST "),a("code",null,"/api/pay/create"),l(" → 返回支付链接或二维码")]),a("li",null,[a("code",null,"pay_scene=h5"),l(" → 使用 "),a("code",null,"pay_url"),l(" 跳转唤起支付")]),a("li",null,[a("code",null,"pay_scene=native"),l(" → 使用 "),a("code",null,"qr_code_url"),l(" 展示二维码")]),a("li",null,"用户完成支付 → 平台异步通知 → 订单状态更新为已支付"),a("li",null,[l("GET "),a("code",null,"/api/pay/query"),l(" 查单确认（勿仅依赖 return_url）")])],-1))]),_:1}),s(u,{shadow:"never",class:"page-card",style:{"margin-top":"16px"}},{header:t(()=>[...e[27]||(e[27]=[l("对接提示",-1)])]),default:t(()=>[e[28]||(e[28]=a("ul",{class:"flow"},[a("li",null,[l("金额单位均为"),a("strong",null,"分"),l("（100 = 1.00 元）")]),a("li",null,[l("兼容别名："),a("code",null,"channel_type"),l("=wechat/alipay；"),a("code",null,"biz_remark"),l("=subject")]),a("li",null,[l("每个代收平台在「平台管理」配置"),a("strong",null,"专属 IP 白名单"),l("，请求须携带 "),a("code",null,"X-App-Key")]),a("li",null,"无需传 store_id、商户号，系统自动从该平台码池轮询收款")],-1))]),_:1})]),_:1})]),_:1})])}}},ue=N(oe,[["__scopeId","data-v-6681576e"]]);export{ue as default};
