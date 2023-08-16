"use strict";(self.webpackChunk=self.webpackChunk||[]).push([[6835],{5273:(n,s,a)=>{a.r(s),a.d(s,{data:()=>e});const e={key:"v-6787fb4c",path:"/guide/features/upload_files.html",title:"Upload files",lang:"en-US",frontmatter:{},excerpt:"",headers:[],filePathRelative:"guide/features/upload_files.md",git:{updatedTime:1667058369e3,contributors:[{name:"Anton",email:"a.shumansky@gmail.com",commits:1}]}}},235:(n,s,a)=>{a.r(s),a.d(s,{default:()=>p});const e=(0,a(6252).uE)('<h1 id="upload-files" tabindex="-1"><a class="header-anchor" href="#upload-files" aria-hidden="true">#</a> Upload files</h1><p>Hitrix allow us to upload files using OSS service and assign them to FileEntity. Whenever we want to reassign them to the right entity we need to do something like that in separate endpoint</p><div class="language-go ext-go line-numbers-mode"><pre class="language-go"><code>    <span class="token comment">//...</span>\n\tfileEntity <span class="token operator">:=</span> <span class="token operator">&amp;</span>entity<span class="token punctuation">.</span>FileEntity<span class="token punctuation">{</span><span class="token punctuation">}</span>\n\tfound <span class="token operator">:=</span> ormService<span class="token punctuation">.</span><span class="token function">LoadByID</span><span class="token punctuation">(</span>fileID<span class="token punctuation">,</span> fileEntity<span class="token punctuation">)</span>\n\n\t<span class="token keyword">if</span> <span class="token operator">!</span>found <span class="token punctuation">{</span>\n\t\t<span class="token keyword">return</span> fmt<span class="token punctuation">.</span><span class="token function">Errorf</span><span class="token punctuation">(</span><span class="token string">&quot;file with FileID %v not found&quot;</span><span class="token punctuation">,</span> <span class="token operator">*</span>fileID<span class="token punctuation">)</span>\n\t<span class="token punctuation">}</span>\n\n\t<span class="token keyword">if</span> fileEntity<span class="token punctuation">.</span>Namespace <span class="token operator">!=</span> oss<span class="token punctuation">.</span>NamespaceUserAvatar<span class="token punctuation">.</span><span class="token function">String</span><span class="token punctuation">(</span><span class="token punctuation">)</span> <span class="token punctuation">{</span>\n\t\t<span class="token keyword">return</span> goErrors<span class="token punctuation">.</span><span class="token function">New</span><span class="token punctuation">(</span><span class="token string">&quot;wrong file category&quot;</span><span class="token punctuation">)</span>\n\t<span class="token punctuation">}</span>\n\n\tuserEntity<span class="token punctuation">.</span>Avatar <span class="token operator">=</span> fileEntity<span class="token punctuation">.</span>File\n    <span class="token comment">//...</span>\n</code></pre><div class="line-numbers"><span class="line-number">1</span><br><span class="line-number">2</span><br><span class="line-number">3</span><br><span class="line-number">4</span><br><span class="line-number">5</span><br><span class="line-number">6</span><br><span class="line-number">7</span><br><span class="line-number">8</span><br><span class="line-number">9</span><br><span class="line-number">10</span><br><span class="line-number">11</span><br><span class="line-number">12</span><br><span class="line-number">13</span><br><span class="line-number">14</span><br></div></div><p>If you want to enable this feature you should call <code>middleware.FileRouter(ginEngine)</code> This will add <code>/v1/file/upload/</code> endpoint where the customers can upload their files</p>',4),t={},p=(0,a(3744).Z)(t,[["render",function(n,s){return e}]])},3744:(n,s)=>{s.Z=(n,s)=>{for(const[a,e]of s)n[a]=e;return n}}}]);