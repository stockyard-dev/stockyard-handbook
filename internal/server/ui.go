package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html>
<html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Handbook</title>
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#c45d2c;--rl:#e8753a;--leather:#a0845c;--ll:#c4a87a;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c44040;--mono:'JetBrains Mono',Consolas,monospace;--serif:'Libre Baskerville',Georgia,serif}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);font-size:13px;line-height:1.6;height:100vh;overflow:hidden}
a{color:var(--rl);text-decoration:none}a:hover{color:var(--gold)}
.app{display:flex;height:100vh}
.sidebar{width:240px;background:var(--bg2);border-right:1px solid var(--bg3);display:flex;flex-direction:column;flex-shrink:0;overflow-y:auto}
.sidebar-hdr{padding:.6rem .8rem;border-bottom:1px solid var(--bg3);font-family:var(--serif);font-size:.9rem;display:flex;justify-content:space-between;align-items:center}
.sidebar-hdr span{color:var(--rl)}
.sidebar-search{padding:.4rem .6rem;border-bottom:1px solid var(--bg3)}
.sidebar-search input{width:100%;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);padding:.3rem .5rem;font-family:var(--mono);font-size:.72rem;outline:none}
.sidebar-search input:focus{border-color:var(--rust)}
.space-hdr{padding:.4rem .8rem;font-size:.65rem;text-transform:uppercase;letter-spacing:1px;color:var(--rust);cursor:pointer;display:flex;justify-content:space-between;align-items:center;margin-top:.3rem}
.space-hdr:hover{color:var(--rl)}
.tree-item{padding:.25rem .8rem;padding-left:1.2rem;font-size:.75rem;cursor:pointer;color:var(--cd);transition:background .1s;display:flex;align-items:center;gap:.3rem;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.tree-item:hover{background:var(--bg3)}.tree-item.active{background:var(--bg3);color:var(--cream)}
.tree-item .child-ind{color:var(--cm);font-size:.6rem}
.tree-children{padding-left:.8rem}

.content{flex:1;display:flex;flex-direction:column;min-width:0}
.content-toolbar{padding:.4rem .8rem;border-bottom:1px solid var(--bg3);display:flex;align-items:center;gap:.5rem}
.btn{font-family:var(--mono);font-size:.68rem;padding:.25rem .6rem;border:1px solid;cursor:pointer;background:transparent;transition:.15s;white-space:nowrap}
.btn-p{border-color:var(--rust);color:var(--rl)}.btn-p:hover{background:var(--rust);color:var(--cream)}
.btn-d{border-color:var(--bg3);color:var(--cm)}.btn-d:hover{border-color:var(--red);color:var(--red)}
.btn-s{border-color:var(--green);color:var(--green)}.btn-s:hover{background:var(--green);color:var(--bg)}

.page-title{width:100%;background:transparent;border:none;color:var(--cream);font-family:var(--serif);font-size:1.2rem;padding:.6rem .8rem;outline:none;border-bottom:1px solid var(--bg3)}
.page-body{flex:1;display:flex;overflow:hidden}
.page-body textarea{flex:1;background:transparent;border:none;color:var(--cd);font-family:var(--mono);font-size:.8rem;padding:.8rem;outline:none;resize:none;line-height:1.7}
.page-view{flex:1;padding:.8rem 1.2rem;overflow-y:auto;font-size:.82rem;color:var(--cd);line-height:1.8}
.page-view h1,.page-view h2,.page-view h3{color:var(--cream);font-family:var(--serif);margin:1rem 0 .4rem}
.page-view h1{font-size:1.3rem}.page-view h2{font-size:1.05rem}.page-view h3{font-size:.9rem}
.page-view code{background:var(--bg3);padding:.1rem .3rem;font-size:.75rem}
.page-view pre{background:var(--bg3);padding:.6rem;margin:.5rem 0;overflow-x:auto;font-size:.75rem}
.page-view pre code{background:transparent;padding:0}
.page-view ul,.page-view ol{padding-left:1.2rem;margin:.4rem 0}
.page-view blockquote{border-left:3px solid var(--rust);padding-left:.8rem;color:var(--cm);margin:.5rem 0}
.page-view p{margin:.4rem 0}

.comments-pane{border-top:1px solid var(--bg3);max-height:200px;overflow-y:auto;padding:.5rem .8rem}
.cmt-item{padding:.3rem 0;border-bottom:1px solid var(--bg3);font-size:.72rem}
.cmt-author{color:var(--rl);font-weight:600}.cmt-time{color:var(--cm);font-size:.6rem;margin-left:.4rem}
.cmt-body{color:var(--cd);margin-top:.1rem}

.empty{display:flex;align-items:center;justify-content:center;flex:1;color:var(--cm);font-style:italic;font-family:var(--serif)}

.modal-bg{position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,.65);display:flex;align-items:center;justify-content:center;z-index:100}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:90%;max-width:400px}
.modal h2{font-family:var(--serif);font-size:.9rem;margin-bottom:.8rem}
label.fl{display:block;font-size:.65rem;color:var(--leather);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem;margin-top:.5rem}
input[type=text],select{background:var(--bg);border:1px solid var(--bg3);color:var(--cream);padding:.35rem .5rem;font-family:var(--mono);font-size:.78rem;width:100%;outline:none}
</style>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital@0;1&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
</head><body>
<div class="app">
<div class="sidebar">
<div class="sidebar-hdr"><span>Handbook</span><button class="btn btn-p" style="font-size:.6rem;padding:.15rem .4rem" onclick="showNewSpace()">+ Space</button></div>
<div class="sidebar-search"><input type="text" id="searchBox" placeholder="Search pages..." onkeydown="if(event.key==='Enter')doSearch()"></div>
<div id="treeContainer"></div>
</div>
<div class="content" id="contentArea">
<div class="empty" id="emptyState">Select a page or create a new one.</div>
</div>
</div>
<div id="modal"></div>

<script>
let spaces=[],curPage=null,curSpace='',editing=false,saveTimer=null;

async function api(url,opts){const r=await fetch(url,opts);return r.json()}
function esc(s){return String(s||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;')}
function timeAgo(d){if(!d)return'';const s=Math.floor((Date.now()-new Date(d))/1e3);if(s<60)return s+'s ago';if(s<3600)return Math.floor(s/60)+'m ago';if(s<86400)return Math.floor(s/3600)+'h ago';return Math.floor(s/86400)+'d ago'}

async function init(){
  const d=await api('/api/spaces');spaces=d.spaces||[];
  renderTree();
}

async function renderTree(){
  let html='';
  for(const sp of spaces){
    const td=await api('/api/spaces/'+sp.id+'/tree');
    html+='<div class="space-hdr" onclick="curSpace=\''+sp.id+'\'">'+esc(sp.icon||'📁')+' '+esc(sp.name)+
      '<span style="display:flex;gap:.2rem"><span style="font-size:.55rem;color:var(--cm)">'+sp.page_count+'</span>'+
      '<span style="cursor:pointer;color:var(--rl);font-size:.7rem" onclick="event.stopPropagation();newPage(\''+sp.id+'\',\'\')">+</span></span></div>';
    html+=renderNodes(td.tree||[],sp.id);
  }
  document.getElementById('treeContainer').innerHTML=html||'<div style="padding:1rem .8rem;font-size:.75rem;color:var(--cm)">No spaces yet.</div>';
}

function renderNodes(nodes,spaceID){
  let html='';
  for(const n of nodes){
    const active=curPage&&curPage.id===n.page.id?'active':'';
    const childInd=n.page.child_count?'<span class="child-ind">'+n.page.child_count+'</span>':'';
    html+='<div class="tree-item '+active+'" onclick="openPage(\''+n.page.id+'\')">'+
      (n.page.status==='draft'?'<span style="color:var(--gold);font-size:.55rem">DRAFT</span> ':'')+
      esc(n.page.title)+childInd+'</div>';
    if(n.children&&n.children.length){
      html+='<div class="tree-children">'+renderNodes(n.children,spaceID)+'</div>';
    }
  }
  return html;
}

async function openPage(id){
  curPage=await api('/api/pages/'+id);
  editing=false;
  renderPage();
  renderTree();
}

function renderPage(){
  if(!curPage){document.getElementById('contentArea').innerHTML='<div class="empty">Select a page.</div>';return}
  const p=curPage;
  if(editing){
    document.getElementById('contentArea').innerHTML=
      '<div class="content-toolbar">'+
        '<button class="btn btn-s" onclick="savePage()">Save</button>'+
        '<button class="btn btn-d" onclick="editing=false;renderPage()">Cancel</button>'+
        '<select id="pgStatus"><option value="published"'+(p.status==='published'?' selected':'')+'>Published</option><option value="draft"'+(p.status==='draft'?' selected':'')+'>Draft</option></select>'+
        '<span style="flex:1"></span><span style="font-size:.6rem;color:var(--cm)">'+p.word_count+'w · '+p.revision_count+' revisions</span>'+
      '</div>'+
      '<input class="page-title" id="pgTitle" value="'+esc(p.title)+'">'+
      '<div class="page-body"><textarea id="pgBody">'+esc(p.body)+'</textarea></div>';
  } else {
    const rendered=renderMd(p.body);
    document.getElementById('contentArea').innerHTML=
      '<div class="content-toolbar">'+
        '<button class="btn btn-p" onclick="editing=true;renderPage()">Edit</button>'+
        '<button class="btn btn-d" onclick="showRevisions(\''+p.id+'\')">History ('+p.revision_count+')</button>'+
        '<button class="btn btn-d" onclick="newPage(\''+p.space_id+'\',\''+p.id+'\')">+ Child</button>'+
        '<span style="flex:1"></span>'+
        '<span style="font-size:.6rem;color:var(--cm)">'+p.word_count+'w · '+timeAgo(p.updated_at)+'</span>'+
        '<button class="btn btn-d" onclick="if(confirm(\'Delete?\'))delPage(\''+p.id+'\')">Del</button>'+
      '</div>'+
      '<div class="page-view"><h1>'+esc(p.title)+'</h1>'+rendered+'</div>'+
      '<div class="comments-pane" id="commentsPane"></div>';
    loadComments(p.id);
  }
}

async function savePage(){
  if(!curPage)return;
  const body={title:document.getElementById('pgTitle').value,body:document.getElementById('pgBody').value,status:document.getElementById('pgStatus').value};
  await api('/api/pages/'+curPage.id,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
  curPage=await api('/api/pages/'+curPage.id);
  editing=false;renderPage();renderTree();
}

async function delPage(id){
  await api('/api/pages/'+id,{method:'DELETE'});curPage=null;
  document.getElementById('contentArea').innerHTML='<div class="empty">Page deleted.</div>';
  init();
}

async function newPage(spaceID,parentID){
  const body={space_id:spaceID,parent_id:parentID||'',title:'New Page',body:'',status:'draft'};
  const p=await api('/api/pages',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
  curPage=p;editing=true;renderPage();init();
}

async function loadComments(pageID){
  const d=await api('/api/pages/'+pageID+'/comments');
  const cmts=(d.comments||[]).map(c=>'<div class="cmt-item"><span class="cmt-author">'+esc(c.author||'anon')+'</span><span class="cmt-time">'+timeAgo(c.created_at)+'</span><div class="cmt-body">'+esc(c.body)+'</div></div>').join('');
  const el=document.getElementById('commentsPane');
  if(el)el.innerHTML='<div style="font-size:.65rem;color:var(--leather);margin-bottom:.3rem">Comments ('+((d.comments||[]).length)+')</div>'+cmts+
    '<div style="display:flex;gap:.3rem;margin-top:.3rem"><input type="text" id="cmtAuthor" placeholder="Name" style="width:80px;font-size:.68rem"><input type="text" id="cmtBody" placeholder="Add comment..." style="flex:1;font-size:.68rem"><button class="btn btn-p" style="font-size:.6rem" onclick="addComment(\''+pageID+'\')">Post</button></div>';
}

async function addComment(pageID){
  const author=document.getElementById('cmtAuthor').value||'anon';
  const body=document.getElementById('cmtBody').value;if(!body)return;
  await api('/api/pages/'+pageID+'/comments',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({author,body})});
  loadComments(pageID);
}

async function showRevisions(pageID){
  const d=await api('/api/pages/'+pageID+'/revisions');
  const revs=(d.revisions||[]).map(r=>'<div style="padding:.3rem 0;border-bottom:1px solid var(--bg3);font-size:.72rem;cursor:pointer" onclick="viewRevision(\''+r.id+'\')">'+
    '<b>'+esc(r.title)+'</b> <span style="color:var(--cm)">by '+esc(r.author||'unknown')+' · '+timeAgo(r.created_at)+'</span></div>').join('');
  document.getElementById('modal').innerHTML='<div class="modal-bg" onclick="if(event.target===this)closeModal()"><div class="modal" style="max-width:500px">'+
    '<h2>Page History</h2>'+(revs||'<div style="color:var(--cm)">No revisions yet.</div>')+
    '<div style="margin-top:.5rem"><button class="btn btn-d" onclick="closeModal()">Close</button></div></div></div>';
}

async function viewRevision(id){
  const r=await api('/api/revisions/'+id);
  document.getElementById('modal').innerHTML='<div class="modal-bg" onclick="if(event.target===this)closeModal()"><div class="modal" style="max-width:600px;max-height:80vh;overflow-y:auto">'+
    '<h2>'+esc(r.title)+'</h2><div style="font-size:.65rem;color:var(--cm)">by '+esc(r.author)+' · '+timeAgo(r.created_at)+'</div>'+
    '<pre style="margin-top:.5rem;padding:.5rem;background:var(--bg);font-size:.72rem;white-space:pre-wrap">'+esc(r.body)+'</pre>'+
    '<button class="btn btn-d" style="margin-top:.5rem" onclick="closeModal()">Close</button></div></div>';
}

async function doSearch(){
  const q=document.getElementById('searchBox').value;if(!q)return;
  const d=await api('/api/search?q='+encodeURIComponent(q));
  const pages=d.pages||[];
  const html=pages.map(p=>'<div class="tree-item" onclick="openPage(\''+p.id+'\')">'+esc(p.title)+'<span style="color:var(--cm);font-size:.6rem;margin-left:.3rem">'+p.word_count+'w</span></div>').join('');
  document.getElementById('treeContainer').innerHTML='<div class="space-hdr">Search: "'+esc(q)+'" ('+pages.length+')</div>'+html+
    '<div class="tree-item" style="color:var(--rl)" onclick="init()">← Back to tree</div>';
}

function showNewSpace(){
  document.getElementById('modal').innerHTML='<div class="modal-bg" onclick="if(event.target===this)closeModal()"><div class="modal">'+
    '<h2>New Space</h2><label class="fl">Name</label><input type="text" id="ns-name">'+
    '<label class="fl">Icon (emoji)</label><input type="text" id="ns-icon" value="📁" style="width:60px">'+
    '<div style="display:flex;gap:.5rem;margin-top:1rem"><button class="btn btn-p" onclick="saveNewSpace()">Create</button><button class="btn btn-d" onclick="closeModal()">Cancel</button></div></div></div>';
}
async function saveNewSpace(){
  const body={name:document.getElementById('ns-name').value,icon:document.getElementById('ns-icon').value};
  if(!body.name){alert('Name required');return}
  await api('/api/spaces',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
  closeModal();init();
}

function renderMd(md){
  let h=esc(md);
  h=h.replace(/^### (.+)$/gm,'<h3>$1</h3>');h=h.replace(/^## (.+)$/gm,'<h2>$1</h2>');h=h.replace(/^# (.+)$/gm,'<h1>$1</h1>');
  h=h.replace(/\*\*(.+?)\*\*/g,'<strong>$1</strong>');h=h.replace(/\*(.+?)\*/g,'<em>$1</em>');
  h=h.replace(/` + "`" + `([^` + "`" + `]+)` + "`" + `/g,'<code>$1</code>');
  h=h.replace(/^&gt; (.+)$/gm,'<blockquote>$1</blockquote>');
  h=h.replace(/^- (.+)$/gm,'<li>$1</li>');
  h=h.replace(/\n\n/g,'</p><p>');
  return '<p>'+h+'</p>';
}

function closeModal(){document.getElementById('modal').innerHTML=''}
init();
fetch('/api/tier').then(r=>r.json()).then(j=>{if(j.tier==='free'){var b=document.getElementById('upgrade-banner');if(b)b.style.display='block'}}).catch(()=>{var b=document.getElementById('upgrade-banner');if(b)b.style.display='block'});
</script></body></html>`
