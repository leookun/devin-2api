package dashboard

const loginPage = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Devin API - 管理面板</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:system-ui,-apple-system,sans-serif;background:#0f1117;color:#e0e0e0;display:flex;justify-content:center;align-items:center;min-height:100vh}
.login-box{background:#1a1d27;border-radius:12px;padding:40px;width:360px;box-shadow:0 4px 24px rgba(0,0,0,.4)}
.login-box h1{font-size:20px;margin-bottom:24px;text-align:center;color:#7c8aff}
.login-box input{width:100%;padding:12px 16px;border:1px solid #333;border-radius:8px;background:#0f1117;color:#e0e0e0;font-size:14px;margin-bottom:16px}
.login-box button{width:100%;padding:12px;border:none;border-radius:8px;background:#7c8aff;color:#fff;font-size:14px;cursor:pointer}
.error{color:#ff6b6b;font-size:13px;text-align:center;margin-top:8px;display:none}
</style>
</head>
<body>
<div class="login-box">
<h1>Devin API 管理面板</h1>
<form id="loginForm">
<input type="password" id="password" placeholder="请输入密码" autofocus>
<button type="submit">登录</button>
</form>
<div class="error" id="err">密码错误</div>
</div>
<script>
document.getElementById('loginForm').addEventListener('submit',async e=>{
e.preventDefault();
const pwd=document.getElementById('password').value;
const res=await fetch('/panel/login',{method:'POST',headers:{'Content-Type':'application/x-www-form-urlencoded'},body:'password='+encodeURIComponent(pwd)});
if(res.ok){location.reload()}else{const el=document.getElementById('err');el.style.display='block';setTimeout(()=>el.style.display='none',2000)}
});
</script>
</body>
</html>`

// dashboardPage is intentionally a single HTML document with client-side filters.
const dashboardPage = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Devin API - 管理面板</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:system-ui,-apple-system,sans-serif;background:#0f1117;color:#e0e0e0;padding:16px 20px 40px}
.container{max-width:1400px;margin:0 auto}
h1{font-size:22px;margin-bottom:16px;color:#7c8aff}
.section{background:#1a1d27;border-radius:12px;padding:20px;margin-bottom:16px}
.section h2{font-size:15px;margin-bottom:12px;color:#7c8aff}
.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(160px,1fr));gap:10px}
.card{background:#222632;border-radius:8px;padding:12px}
.card .label{font-size:11px;color:#888;margin-bottom:4px}
.card .value{font-size:16px;font-weight:600;word-break:break-all}
.progress-bar{width:100%;height:8px;background:#333;border-radius:4px;margin-top:8px;overflow:hidden}
.progress-fill{height:100%;border-radius:4px;transition:width .5s}
.kv-row{display:flex;flex-wrap:wrap;gap:6px 20px;margin-top:10px}
.kv{font-size:12px;line-height:1.8}
.kv .k{color:#888;margin-right:4px}
.kv .v{color:#e0e0e0;font-weight:600}
.ring-row{display:flex;gap:32px;flex-wrap:wrap;margin-top:14px}
.ring-wrap{display:flex;flex-direction:column;align-items:center;gap:6px}
.ring-wrap .ring-label{font-size:12px;color:#888}
.ring-wrap .ring-pct{font-size:16px;font-weight:700}
.ring-countdown{font-size:11px;color:#666;margin-top:2px}
.ring-svg{width:80px;height:80px;transform:rotate(-90deg)}
.ring-bg{fill:none;stroke:#333;stroke-width:8}
.ring-fg{fill:none;stroke-width:8;stroke-linecap:round;transition:stroke-dashoffset .5s}
.filters{display:flex;flex-wrap:wrap;gap:8px;margin-bottom:12px;align-items:center}
.filters select,.filters input[type=search]{padding:8px 12px;border:1px solid #333;border-radius:8px;background:#0f1117;color:#e0e0e0;font-size:13px}
.filters input[type=search]{flex:1;min-width:180px}
.chip-row{display:flex;flex-wrap:wrap;gap:6px;margin-bottom:10px}
.chip{padding:4px 10px;border-radius:999px;border:1px solid #333;background:#222632;color:#bbb;font-size:12px;cursor:pointer;user-select:none}
.chip:hover{border-color:#7c8aff;color:#fff}
.chip.on{background:#7c8aff22;border-color:#7c8aff;color:#aab4ff}
.stats{font-size:12px;color:#888;margin-bottom:8px}
.model-scroll{max-height:70vh;overflow:auto;border:1px solid #222;border-radius:8px}
table{width:100%;border-collapse:collapse;font-size:12px}
th{text-align:left;padding:8px 10px;border-bottom:2px solid #333;color:#888;font-weight:600;position:sticky;top:0;background:#1a1d27;z-index:1;white-space:nowrap}
td{padding:7px 10px;border-bottom:1px solid #222;vertical-align:top}
tr:hover{background:#222632}
.mono{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:11px}
.badge{display:inline-block;padding:1px 7px;border-radius:4px;font-size:10px;font-weight:600;margin:1px 2px 1px 0;white-space:nowrap}
.badge-free{background:#1a4a2e;color:#4ade80}
.badge-low{background:#1a3a4a;color:#60a5fa}
.badge-medium{background:#4a3a1a;color:#fbbf24}
.badge-high{background:#4a1a1a;color:#f87171}
.badge-promo{background:#4a1a4a;color:#c084fc}
.badge-beta{background:#3a3a1a;color:#fde047}
.badge-fast{background:#1a4a4a;color:#22d3ee}
.badge-img{background:#2a2a3a;color:#a78bfa}
.badge-new{background:#1a3a2a;color:#34d399}
.badge-premium{background:#3a2a1a;color:#f59e0b}
.badge-rec{background:#1a2a4a;color:#93c5fd}
.badge-cap{background:#3a1a2a;color:#f472b6}
.badge-off{background:#333;color:#999}
.price{font-variant-numeric:tabular-nums;white-space:nowrap}
.muted{color:#666}
.loading{text-align:center;padding:32px;color:#888}
.note{font-size:12px;color:#888;line-height:1.55;margin-top:10px}
.note code{background:#222;padding:1px 5px;border-radius:3px}
</style>
</head>
<body>
<div class="container">
<h1>Devin API 管理面板</h1>

<div class="section" id="statusSection">
<h2>账户 / 容量状态</h2>
<div class="loading">加载中...</div>
</div>

<div class="section">
<h2>模型列表 · 价格 · 筛选</h2>
<div class="filters">
<input type="search" id="search" placeholder="搜索 UID / 标签 / 描述 / 系列..." oninput="applyFilters()">
<select id="fProvider" onchange="applyFilters()"><option value="">全部渠道</option></select>
<select id="fApi" onchange="applyFilters()"><option value="">全部 API Provider</option></select>
<select id="fTier" onchange="applyFilters()">
<option value="">全部定价等级</option>
<option value="free">FREE</option>
<option value="low">LOW</option>
<option value="medium">MEDIUM</option>
<option value="high">HIGH</option>
</select>
<select id="fPricing" onchange="applyFilters()"><option value="">全部计费类型</option></select>
<select id="fSort" onchange="applyFilters()">
<option value="default">默认排序</option>
<option value="mult_asc">倍率 升序</option>
<option value="mult_desc">倍率 降序</option>
<option value="in_asc">Input $/1M 升序</option>
<option value="in_desc">Input $/1M 降序</option>
<option value="out_asc">Output $/1M 升序</option>
<option value="out_desc">Output $/1M 降序</option>
<option value="name">名称 A-Z</option>
</select>
</div>
<div class="chip-row" id="chips">
<span class="chip" data-tag="free" onclick="toggleChip(this)">FREE</span>
<span class="chip" data-tag="promo" onclick="toggleChip(this)">PROMO</span>
<span class="chip" data-tag="img" onclick="toggleChip(this)">支持图片</span>
<span class="chip" data-tag="beta" onclick="toggleChip(this)">BETA</span>
<span class="chip" data-tag="new" onclick="toggleChip(this)">NEW</span>
<span class="chip" data-tag="fast" onclick="toggleChip(this)">FAST</span>
<span class="chip" data-tag="premium" onclick="toggleChip(this)">Premium</span>
<span class="chip" data-tag="rec" onclick="toggleChip(this)">推荐</span>
<span class="chip" data-tag="empty_mult" onclick="toggleChip(this)">无倍率(FREE/基准)</span>
<span class="chip" data-tag="disabled" onclick="toggleChip(this)">已禁用</span>
</div>
<div class="stats" id="stats">加载中...</div>
<div class="model-scroll">
<table id="modelTable">
<thead>
<tr>
<th>#</th>
<th>Model</th>
<th>渠道</th>
<th>等级</th>
<th>倍率</th>
<th>Input $/1M</th>
<th>Cached $/1M</th>
<th>Output $/1M</th>
<th>计费</th>
<th>标签</th>
</tr>
</thead>
<tbody><tr><td colspan="10" class="loading">加载中...</td></tr></tbody>
</table>
</div>
<div class="note">
<strong>价格说明：</strong>
<code>credit_multiplier</code> 是 Windsurf credit 消耗倍率。FREE 模型倍率为 0 表示不扣分；其它模型若上游未单独下发倍率，通常按基准 1.0 理解。
<code>Input / Cached / Output</code> 来自 <code>model_dimensions</code>，单位通常是 <strong>$ / 1M tokens</strong>；上游还可能提供 min~max（不同 effort 区间）。
<code>PROMO</code> 表示促销中。计费类型：STATIC_CREDIT=固定 credit、API=按 API、BYOK=自带 Key、ACU_*=ACU 计费。
</div>
</div>
</div>

<script>
let allModels=[];
let activeTags=new Set();

function esc(s){
return String(s==null?'':s)
.replace(/&/g,'&amp;')
.replace(/</g,'&lt;')
.replace(/>/g,'&gt;')
.replace(/"/g,'&quot;');
}
function card(label,value){
return '<div class="card"><div class="label">'+esc(label)+'</div><div class="value">'+esc(String(value))+'</div></div>';
}
// kv 生成一行键值对（无卡片背景，紧凑排列）
function kv(k,v){return '<span class="kv"><span class="k">'+esc(k)+'</span><span class="v">'+esc(String(v))+'</span></span>'}
// ring 生成环形进度图（SVG），percent 0-100，即使 0 也显示
// resetUnix 为重置时间戳（秒），total 为估算总量（如 monthly_prompt_credits）
function ring(label,percent,resetUnix,total){
const p=Math.max(0,Math.min(100,Number(percent)||0));
const r=32,circ=2*Math.PI*r,off=circ*(1-p/100);
const color=p>50?'#4ade80':p>20?'#fbbf24':'#f87171';
let countdown='';
if(resetUnix&&Number(resetUnix)>0){
const diff=Number(resetUnix)*1000-Date.now();
if(diff>0){
const d=Math.floor(diff/86400000),h=Math.floor(diff%86400000/3600000),m=Math.floor(diff%3600000/60000);
if(d>0){countdown=d+'d '+h+'h '+m+'m'}else if(h>0){countdown=h+'h '+m+'m'}else{countdown=m+'m'}
}else{countdown='已重置'}
}
let actual='-';
if(Number.isFinite(Number(total))&&Number(total)>0){
const rem=Math.round(p/100*Number(total));
const used=Number(total)-rem;
actual=used+' / '+total;
}
return '<div class="ring-wrap"><svg class="ring-svg" viewBox="0 0 80 80"><circle class="ring-bg" cx="40" cy="40" r="'+r+'"/><circle class="ring-fg" cx="40" cy="40" r="'+r+'" stroke="'+color+'" stroke-dasharray="'+circ.toFixed(1)+'" stroke-dashoffset="'+off.toFixed(1)+'"/></svg><div class="ring-pct" style="color:'+color+'">'+p+'%</div><div class="ring-label">'+esc(label)+'</div>'+(countdown?'<div class="ring-countdown">倒计时 '+countdown+'</div>':'')+'</div>';
}

function fmtQuota(v){
if(v==null||v===''||v===undefined) return '-';
if(Number(v)===-1) return '不限 / 配额制';
return String(v);
}
function fmtUnix(v){
if(v==null||v===''||Number(v)===0) return '-';
const n=Number(v);
if(!Number.isFinite(n)||n<=0) return String(v);
try{return new Date(n*1000).toLocaleString()}catch(e){return String(v)}
}
function bar(label,percent){
const p=Number(percent);
if(!Number.isFinite(p)) return '';
if(p<=0){
return '<div style="margin-top:12px"><div class="label" style="font-size:12px;color:#888;margin-bottom:4px">'+esc(label)+': 已用尽</div><div class="progress-bar"><div class="progress-fill" style="width:100%;background:#f87171"></div></div></div>';
}
const color=p>50?'#4ade80':p>20?'#fbbf24':'#f87171';
return '<div style="margin-top:12px"><div class="label" style="font-size:12px;color:#888;margin-bottom:4px">'+esc(label)+': '+p+'%</div><div class="progress-bar"><div class="progress-fill" style="width:'+Math.min(100,p)+'%;background:'+color+'"></div></div></div>';
}

async function loadStatus(){
const res=await fetch('/panel/api/status');
const data=await res.json();
const el=document.getElementById('statusSection');
let html='<h2>账户 / 容量状态</h2>';
if(data.user){
const u=data.user;
html+='<div class="kv-row">';
html+=kv('用户名',u.name||'-');
html+=kv('邮箱',u.email||'-');
html+=kv('Pro',u.pro?'是':'否');
html+=kv('Tier',u.teams_tier||'-');
html+=kv('User ID',u.user_id||'-');
html+='</div>';
}
if(data.plan_status||data.plan_info){
const ps=data.plan_status||{};
const pi=data.plan_info||{};
html+='<h2 style="margin-top:16px">套餐与用量</h2>';
// 环形图：日/周配额剩余（始终显示，即使 0%）
html+='<div class="ring-row">';
const total=ps.monthly_prompt_credits??pi.monthly_prompt_credits;
html+=ring('日配额剩余',ps.daily_quota_remaining!=null?ps.daily_quota_remaining:0,ps.daily_quota_reset,total);
html+=ring('周配额剩余',ps.weekly_quota_remaining!=null?ps.weekly_quota_remaining:0,ps.weekly_quota_reset,total);
html+='</div>';
// 紧凑键值对
html+='<div class="kv-row">';
html+=kv('套餐',ps.plan_name||pi.plan_name||'-');
html+=kv('计费',ps.billing_strategy||pi.billing_strategy||'-');
html+=kv('月 Prompt',fmtQuota(ps.monthly_prompt_credits??pi.monthly_prompt_credits));
html+=kv('可用 Prompt',fmtQuota(ps.available_prompt_credits));
html+=kv('可用 Flow',fmtQuota(ps.available_flow_credits));
html+=kv('可用 Flex',fmtQuota(ps.available_flex_credits));
html+=kv('周期',ps.plan_start||'-'+' ~ '+(ps.plan_end||'-'));
html+=kv('超额',ps.overage_balance_micros??'-');
html+=kv('ACU',(ps.acu_consumed??'-')+' / '+(ps.acu_limit??'-'));
html+='</div>';
html+='<div class="note" style="margin-top:10px">Pro 多为 <strong>配额制 (QUOTA)</strong>：优先看日/周剩余百分比。月 Prompt 为 -1 表示不按固定 monthly credit 计。</div>';
}
if(data.capacity){
html+='<div class="kv-row" style="margin-top:10px">';
html+=kv('有容量',data.capacity.has_capacity?'是':'否');
html+=kv('活跃会话',data.capacity.active_sessions??'-');
html+=kv('容量消息',data.capacity.message||'-');
html+='</div>';
}
if(data.ide_status){
html+='<div class="kv-row" style="margin-top:6px">';
html+=kv('IDE 状态',data.ide_status.level||'-');
html+=kv('IDE 消息',data.ide_status.message||'-');
html+='</div>';
}
if(data.model_statuses&&data.model_statuses.length){
html+='<h2 style="margin-top:16px">模型状态告警</h2><div class="grid">';
data.model_statuses.forEach(s=>{html+=card(s.model||'-',s.status||'-')});
html+='</div>';
}
if(!data.user && data.user_status_error){
html+='<div style="color:#ff6b6b;margin-top:12px;font-size:13px">账户用量拉取失败: '+esc(data.user_status_error)+'</div>';
} else if(!data.user){
html+='<div class="note" style="margin-top:12px">暂无账户用量数据。</div>';
}
if(data.capacity_error){html+='<div style="color:#ff6b6b;margin-top:8px;font-size:12px">'+esc(data.capacity_error)+'</div>'}
el.innerHTML=html;
}

function toggleChip(el){
const tag=el.dataset.tag;
if(activeTags.has(tag)){activeTags.delete(tag);el.classList.remove('on')}
else{activeTags.add(tag);el.classList.add('on')}
applyFilters();
}

function fillSelect(id,values){
const el=document.getElementById(id);
const cur=el.value;
const keep=el.options[0].outerHTML;
const opts=[...values].filter(Boolean).sort().map(v=>'<option value="'+esc(v)+'">'+esc(v)+'</option>').join('');
el.innerHTML=keep+opts;
if([...el.options].some(o=>o.value===cur)) el.value=cur;
}

function multDisplay(m){
if(m.cost_tier==='free' && (!m.credit_multiplier || m.credit_multiplier===0)){
return '<span class="badge badge-free">0 (FREE)</span>';
}
if(!m.multiplier_known || m.credit_multiplier===0){
return '<span class="muted" title="上游未单独下发倍率，通常按基准 1.0">— / ≈1.0</span>';
}
const n=Number(m.credit_multiplier);
return 'x'+(n%1?n.toFixed(1):String(n));
}

function money(v){
if(v==null||v===undefined||v==='') return '<span class="muted">—</span>';
const n=Number(v);
if(Number.isNaN(n)) return '<span class="muted">—</span>';
return '<span class="price">$'+n.toFixed(n>=10?1:n>=1?2:3)+'</span>';
}

function badges(m){
let b='';
if(m.cost_tier==='free') b+='<span class="badge badge-free">FREE</span>';
else if(m.cost_tier==='low') b+='<span class="badge badge-low">LOW</span>';
else if(m.cost_tier==='medium') b+='<span class="badge badge-medium">MEDIUM</span>';
else if(m.cost_tier==='high') b+='<span class="badge badge-high">HIGH</span>';
if(m.promo&&m.promo.active) b+='<span class="badge badge-promo">PROMO'+(m.promo.label?(' · '+esc(m.promo.label)):'')+'</span>';
if(m.is_beta) b+='<span class="badge badge-beta">BETA</span>';
if(m.is_new) b+='<span class="badge badge-new">NEW</span>';
if(m.fast&&m.fast.active) b+='<span class="badge badge-fast">FAST</span>';
if(m.supports_images) b+='<span class="badge badge-img">img</span>';
if(m.is_premium) b+='<span class="badge badge-premium">Premium</span>';
if(m.is_recommended) b+='<span class="badge badge-rec">推荐</span>';
if(m.is_capacity_limited) b+='<span class="badge badge-cap">限容</span>';
if(m.disabled) b+='<span class="badge badge-off">禁用</span>';
return b;
}

function matchTags(m){
for(const t of activeTags){
if(t==='free' && m.cost_tier!=='free') return false;
if(t==='promo' && !(m.promo&&m.promo.active)) return false;
if(t==='img' && !m.supports_images) return false;
if(t==='beta' && !m.is_beta) return false;
if(t==='new' && !m.is_new) return false;
if(t==='fast' && !(m.fast&&m.fast.active)) return false;
if(t==='premium' && !m.is_premium) return false;
if(t==='rec' && !m.is_recommended) return false;
if(t==='empty_mult'){
const empty=!m.multiplier_known || m.credit_multiplier===0;
if(!empty) return false;
}
if(t==='disabled' && !m.disabled) return false;
}
return true;
}

function multOf(m){
if(m.cost_tier==='free') return 0;
if(!m.multiplier_known || !m.credit_multiplier) return 1;
return Number(m.credit_multiplier);
}

function applyFilters(){
const q=document.getElementById('search').value.trim().toLowerCase();
const provider=document.getElementById('fProvider').value;
const api=document.getElementById('fApi').value;
const tier=document.getElementById('fTier').value;
const pricing=document.getElementById('fPricing').value;
const sort=document.getElementById('fSort').value;

let list=allModels.filter(m=>{
if(provider && m.provider!==provider) return false;
if(api && m.api_provider!==api) return false;
if(tier && m.cost_tier!==tier) return false;
if(pricing && m.pricing_type!==pricing) return false;
if(!matchTags(m)) return false;
if(q){
const hay=[m.uid,m.label,m.description,m.family,m.provider,m.api_provider].join(' ').toLowerCase();
if(!hay.includes(q)) return false;
}
return true;
});

list=list.slice().sort((a,b)=>{
switch(sort){
case 'mult_asc': return multOf(a)-multOf(b);
case 'mult_desc': return multOf(b)-multOf(a);
case 'in_asc': return (a.price_input??1e9)-(b.price_input??1e9);
case 'in_desc': return (b.price_input??-1)-(a.price_input??-1);
case 'out_asc': return (a.price_output??1e9)-(b.price_output??1e9);
case 'out_desc': return (b.price_output??-1)-(a.price_output??-1);
case 'name': return String(a.label||a.uid).localeCompare(String(b.label||b.uid));
default: return 0;
}
});
renderModels(list);
}

function renderModels(models){
const tbody=document.querySelector('#modelTable tbody');
document.getElementById('stats').textContent='显示 '+models.length+' / 共 '+allModels.length+' 个模型';
if(!models.length){
tbody.innerHTML='<tr><td colspan="10" class="loading">无匹配模型</td></tr>';
return;
}
let html='';
models.forEach((m,i)=>{
const dimTip=(m.dimensions||[]).map(d=>d.label+': '+d.value+(d.min||d.max?(' (min '+d.min+' ~ max '+d.max+')'):'')+' / '+(d.denominator||'')).join(' | ');
const title=[m.description,m.family?('系列: '+m.family):'',dimTip,m.beta_warning||''].filter(Boolean).join(' | ');
html+='<tr title="'+esc(title)+'">';
html+='<td>'+(i+1)+'</td>';
html+='<td><div><strong>'+esc(m.label||'-')+'</strong></div><div class="mono muted">'+esc(m.uid)+'</div></td>';
html+='<td>'+esc(m.provider||'-');
if(m.api_provider && m.api_provider!==m.provider){
html+='<div class="muted mono">'+esc(m.api_provider)+'</div>';
}
html+='</td>';
html+='<td>'+esc(m.cost_tier||'-')+'</td>';
html+='<td>'+multDisplay(m)+'</td>';
html+='<td>'+money(m.price_input)+'</td>';
html+='<td>'+money(m.price_cached)+'</td>';
html+='<td>'+money(m.price_output)+'</td>';
html+='<td class="mono">'+esc(m.pricing_type||'-')+'</td>';
html+='<td>'+badges(m)+'</td>';
html+='</tr>';
});
tbody.innerHTML=html;
}

async function loadModels(){
const res=await fetch('/panel/api/models');
const data=await res.json();
allModels=data.models||[];
const providers=new Set(), apis=new Set(), pricings=new Set();
allModels.forEach(m=>{
if(m.provider) providers.add(m.provider);
if(m.api_provider) apis.add(m.api_provider);
if(m.pricing_type) pricings.add(m.pricing_type);
});
fillSelect('fProvider',providers);
fillSelect('fApi',apis);
fillSelect('fPricing',pricings);
applyFilters();
}

loadStatus();
loadModels();
</script>
</body>
</html>`
