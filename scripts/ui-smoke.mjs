// Browser smoke checks against the production bundle and a deterministic Wails
// bridge. No packages are installed. Requires Node 22+ and Microsoft Edge.
import { createServer } from 'node:http';
import { readFile, mkdir, mkdtemp, rm, writeFile } from 'node:fs/promises';
import { resolve, join, extname } from 'node:path';
import { tmpdir } from 'node:os';
import { spawn } from 'node:child_process';
import assert from 'node:assert/strict';

const root = resolve('frontend/dist');
const screenshots = resolve('build/ui-review');
await mkdir(screenshots, { recursive: true });
const profile = await mkdtemp(join(tmpdir(), 'rewinder-ui-'));
const server = createServer(async (req, res) => {
 try {
  const path = resolve(root, '.' + new URL(req.url, 'http://localhost').pathname);
  if (path !== root && !path.startsWith(root + '\\') && !path.startsWith(root + '/')) {res.writeHead(403).end();return;}
  const file = path === root ? join(root, 'index.html') : path;
  res.setHeader('Content-Type', ({'.js':'application/javascript','.css':'text/css','.html':'text/html'})[extname(file)] ?? 'application/octet-stream');
  res.end(await readFile(file));
 } catch {res.writeHead(404).end();}
});
await new Promise(r => server.listen(0, '127.0.0.1', r));
const edge = spawn(process.env.EDGE_PATH ?? 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe', [
 '--headless=new', '--disable-gpu', '--no-first-run', '--no-default-browser-check', '--remote-debugging-port=0', `--user-data-dir=${profile}`, 'about:blank',
], {windowsHide:true, stdio:'ignore'});
const sleep = ms => new Promise(r => setTimeout(r,ms));
let ws;
try {
 let port;
 for(let n=0;n<100;n++) {try{port=Number((await readFile(join(profile,'DevToolsActivePort'),'utf8')).split('\n')[0]);break;}catch{await sleep(100);}}
 assert(port,'Edge did not start');
 const tabs=await (await fetch(`http://127.0.0.1:${port}/json/list`)).json();
 ws=new WebSocket(tabs.find(t=>t.type==='page').webSocketDebuggerUrl);
 await new Promise((r,j)=>{ws.onopen=r;ws.onerror=j;});
 let serial=0; const pending=new Map(); const exceptions=[];
 ws.onmessage=e=>{const m=JSON.parse(e.data);if(m.id){const p=pending.get(m.id);pending.delete(m.id);if(m.error)p?.reject(m.error);else p?.resolve(m.result);}else if(m.method==='Runtime.exceptionThrown'){exceptions.push(m.params.exceptionDetails);}};
 const send=(method,params={})=>new Promise((resolve,reject)=>{const id=++serial;pending.set(id,{resolve,reject});ws.send(JSON.stringify({id,method,params}));});
 const evaluate=async expression=>{const r=await send('Runtime.evaluate',{expression,returnByValue:true,awaitPromise:true});if(r.exceptionDetails)throw new Error(JSON.stringify(r.exceptionDetails));return r.result.value;};
 const waitFor=async expression=>{for(let n=0;n<100;n++){if(await evaluate(expression))return;await sleep(50);}throw new Error(`Timed out: ${expression}`);};
 const click=async text=>{assert(await evaluate(`(()=>{const el=[...document.querySelectorAll('button')].find(b=>b.textContent.trim()===${JSON.stringify(text)});el?.click();return !!el;})()`),`Button missing: ${text}`);await sleep(250);};
 const key=async (key,code=key,modifiers=0)=>{await send('Input.dispatchKeyEvent',{type:'keyDown',key,code,modifiers});await send('Input.dispatchKeyEvent',{type:'keyUp',key,code,modifiers:0});await sleep(220);};
 const screenshot=async name=>{const r=await send('Page.captureScreenshot',{format:'png'});await writeFile(join(screenshots,`${name}.png`),Buffer.from(r.data,'base64'));};
 await send('Runtime.enable');await send('Page.enable');
 await send('Emulation.setDeviceMetricsOverride',{width:1024,height:768,deviceScaleFactor:1,mobile:false});
 await send('Page.addScriptToEvaluateOnNewDocument',{source:`
  window.events={};window.calls=[];
  window.activeRAF=new Set();window.activeTimers=new Set();
  const raf=window.requestAnimationFrame.bind(window),cancel=window.cancelAnimationFrame.bind(window);
  window.requestAnimationFrame=cb=>{const id=raf(t=>{window.activeRAF.delete(id);cb(t);});window.activeRAF.add(id);return id;};
  window.cancelAnimationFrame=id=>{window.activeRAF.delete(id);cancel(id);};
  const timeout=window.setTimeout.bind(window),clear=window.clearTimeout.bind(window);
  window.setTimeout=(cb,ms,...args)=>{const id=timeout(()=>{window.activeTimers.delete(id);cb(...args);},ms);window.activeTimers.add(id);return id;};
  window.clearTimeout=id=>{window.activeTimers.delete(id);clear(id);};
  window.emit=(name,p)=>{for(const cb of window.events[name]??[])cb(p);};
  window.runtime={EventsOnMultiple(name,cb){(window.events[name]??=[]).push(cb);return()=>window.events[name]=window.events[name].filter(x=>x!==cb);}};
  const preferences={hotkey:{mods:3,key:90},language:'en',theme:{accent:'#a99bff',radius:14,density:'normal',motion:'normal',background:{kind:'color',color:'#0a0a0c',endColor:'#171126',image:'',dim:.65,blur:0}}};
  const applications=Array.from({length:70},(_,i)=>({appID:'app-'+i,name:i===0?'Explorer':'Application '+i,executablePath:'C:/app.exe',lastActivityUTC:Date.now(),snapshotCount:180,trackingState:'active',retentionStatus:'ok'}));
  const history=id=>Array.from({length:180},(_,i)=>({appID:id,snapshotID:id+'-'+i,timestampUTC:Date.now()-i*60000,windowsCount:3,filesAdded:1,filesRemoved:0}));
  window.testSettings=preferences;window.testHistory=history;
  window.go={main:{App:{
   GetApps:async()=>structuredClone(applications),GetAppSummary:async id=>structuredClone(applications.find(a=>a.appID===id)),GetTimeline:async id=>history(id),
   GetWindowVisible:async()=>true,GetSettings:async()=>structuredClone(preferences),GetAppInfo:async()=>({version:'0.2.0',systemLanguage:'en',logPath:'C:/Rewinder/rewinder.log',historyPath:'C:/Rewinder/history'}),GetBackgroundImage:async()=>'',
   SaveSettings:async s=>{window.calls.push(['SaveSettings',structuredClone(s)]);Object.assign(preferences,structuredClone(s));},
   SetHotkey:async(mods,key)=>{window.calls.push(['SetHotkey',mods,key]);if(key===88)throw 'hotkey.taken';preferences.hotkey={mods,key};},
   GetAutostart:async()=>false,SetAutostart:async()=>{},PauseTracking:async id=>window.emit('onTrackingStateChanged',{appID:id,state:'paused'}),ResumeTracking:async id=>window.emit('onTrackingStateChanged',{appID:id,state:'active'}),
   LogUI:async s=>window.calls.push(['LogUI',s]),HideOverlay:async()=>{window.calls.push(['HideOverlay']);window.emit('onWindowVisibilityChanged',false);},
   Restore:async(id,snap)=>window.calls.push(['Restore',id,snap]),ShowTimelineWindow:async()=>{window.calls.push(['ShowTimelineWindow']);window.emit('onShowTimelineRequested');},
   HideWindow:async()=>window.emit('onWindowVisibilityChanged',false),MinimiseWindow:async()=>window.emit('onWindowVisibilityChanged',false),ExitApp:async()=>{},
   PickBackgroundImage:async()=>{const c=document.createElement('canvas');c.width=320;c.height=180;const ctx=c.getContext('2d');ctx.fillStyle='#5274b7';ctx.fillRect(0,0,320,180);return c.toDataURL();},
  }}};
 `});
 await send('Page.navigate',{url:`http://127.0.0.1:${server.address().port}/`});
 await waitFor(`document.querySelectorAll('button.snapshot').length===30`);
 assert.equal(await evaluate(`document.querySelectorAll('button.app').length`),40);
 await screenshot('timeline');
 await evaluate(`document.querySelector('button[aria-label="Settings"]').click()`);await waitFor(`document.querySelector('[aria-modal="true"]')`);
 await key('Tab');assert(await evaluate(`document.querySelector('[aria-modal="true"]').contains(document.activeElement)`));
 await click('Appearance');
 await evaluate(`(()=>{const s=document.querySelector('input[type="range"]');s.value='22';s.dispatchEvent(new Event('input',{bubbles:true}));s.dispatchEvent(new Event('change',{bubbles:true}));})()`);
 await waitFor(`window.testSettings.theme.radius===22`);
 assert.equal(await evaluate(`document.documentElement.style.getPropertyValue('--radius')`),'22px');
 await click('Image');await click('Choose image');await waitFor(`document.documentElement.style.getPropertyValue('--background-image').startsWith('url')`);
 await screenshot('appearance');
 await click('Hotkeys');await evaluate(`document.querySelector('input[readonly]').focus()`);await key('x','KeyX',3);
 await waitFor(`document.body.textContent.includes('already used')`);
 assert.equal(await evaluate(`window.testSettings.hotkey.key`),90);
 await evaluate(`document.querySelector('input[readonly]').focus()`);await key('q','KeyQ',3);await waitFor(`window.testSettings.hotkey.key===81`);
 await click('Language');await evaluate(`(()=>{const s=document.querySelector('select');s.value='ru';s.dispatchEvent(new Event('change',{bubbles:true}));})()`);
 await waitFor(`document.documentElement.lang==='ru'`);await screenshot('russian');
 await key('Escape');await waitFor(`!document.querySelector('[aria-modal="true"]')`);
 await evaluate(`document.querySelector('button[aria-label="Настройки"]').click()`);await waitFor(`document.querySelector('[aria-modal="true"]')`);
 await key('Tab','Tab',8);assert(await evaluate(`document.querySelector('[aria-modal="true"]').contains(document.activeElement)`));
 await evaluate(`document.querySelector('.backdrop').click()`);await waitFor(`!document.querySelector('[aria-modal="true"]')`);
 await evaluate(`window.emit('onOverlayOpenRequested',{appID:'app-0'})`);
 await send('Emulation.setDeviceMetricsOverride',{width:760,height:240,deviceScaleFactor:1,mobile:false});
 await waitFor(`document.querySelector('.overlay time')`);await screenshot('overlay');
 await key('ArrowRight');await key('Enter');
 await waitFor(`window.calls.some(c=>c[0]==='Restore')`);
 const order=await evaluate(`window.calls.filter(c=>['Restore','HideOverlay'].includes(c[0])).map(c=>c[0])`);
 assert.deepEqual(order,['HideOverlay','Restore']);
 await waitFor(`!document.querySelector('.shell')&&!document.querySelector('.overlay')`);
 await sleep(250);
 assert.equal(await evaluate(`window.activeRAF.size`),0,'Hidden window still has animation frames');
 assert.equal(await evaluate(`window.activeTimers.size`),0,'Hidden window still has timers');
 await evaluate(`window.hiddenMutations=0;new MutationObserver(list=>window.hiddenMutations+=list.length).observe(document.getElementById('app'),{subtree:true,attributes:true,childList:true,characterData:true})`);
 await evaluate(`window.emit('onSnapshotCreated',{appID:'app-0',snapshot:window.testHistory('app-0')[0]})`);
 await sleep(100);assert.equal(await evaluate(`window.hiddenMutations`),0,'Snapshot updated hidden DOM');
 assert.equal(await evaluate(`document.querySelectorAll('button.snapshot').length`),0);
 await evaluate(`window.emit('onOverlayOpenRequested',{appID:'app-0'});window.emit('onWindowVisibilityChanged',true)`);
 await waitFor(`document.querySelector('.overlay')`);await click('Открыть Rewinder');await waitFor(`document.querySelector('.shell')`);
 await send('Emulation.setDeviceMetricsOverride',{width:760,height:500,deviceScaleFactor:1,mobile:false});await screenshot('minimum-window');
 assert.equal(await evaluate(`document.documentElement.scrollWidth>innerWidth`),false);
 await send('Emulation.setEmulatedMedia',{features:[{name:'prefers-reduced-motion',value:'reduce'}]});
 await evaluate(`document.querySelector('button[aria-label="Настройки"]').click()`);await waitFor(`document.querySelector('[aria-modal="true"]')`);
 await sleep(100);assert.equal(await evaluate(`document.getAnimations().filter(a=>a.playState==='running').length`),0,'Reduced motion still animates');
 await key('Escape');
 assert.deepEqual(exceptions,[]);
 console.log('PASS: lazy lists, modal focus/Esc/backdrop, theme persistence, background, hotkey conflict/success, ru/en, overlay keys, hide-before-restore, hidden DOM, zero hidden timers/RAF, open timeline, reduced motion.');
 console.log('Screenshots:',screenshots);
} finally {
 ws?.close();edge.kill();server.close();
 await sleep(300);
 // This directory is uniquely created by this invocation in the OS temp dir.
 if(profile.startsWith(join(tmpdir(),'rewinder-ui-')))await rm(profile,{recursive:true,force:true,maxRetries:5,retryDelay:200}).catch(()=>{});
}
