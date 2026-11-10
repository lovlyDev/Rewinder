// Solve cubic-bezier(.2,.8,.2,1), matching the CSS --ease token.
export function easing(x: number): number {
 const curve=(t:number,a:number,b:number)=>3*(1-t)*(1-t)*t*a+3*(1-t)*t*t*b+t*t*t;
 let low=0,high=1;
 for(let i=0;i<16;i++){const middle=(low+high)/2;if(curve(middle,.2,.2)<x)low=middle;else high=middle;}
 return curve((low+high)/2,.8,1);
}
export function duration(): number {
 if (document.documentElement.dataset.motion === 'off' || matchMedia('(prefers-reduced-motion: reduce)').matches || document.documentElement.classList.contains('hidden')) return 0;
 return document.documentElement.dataset.motion === 'fast' ? 120 : 180;
}
export function reveal(_node: Element, { scale = false } = {}) {
 return { get duration() {return duration();}, easing, css: (t: number) => `opacity:${t};transform:${scale ? `scale(${.96 + .04*t})` : `translateY(${(1-t)*6}px)`}` };
}
export function fade(_node: Element, _params = {}) {return {get duration(){return duration();},easing,css:(t:number)=>`opacity:${t}`};}
