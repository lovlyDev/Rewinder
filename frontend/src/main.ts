import './theme/fonts.css';
import './theme/tokens.css';
import './theme/global.css';
import { mount } from 'svelte';
import App from './App.svelte';

export default mount(App, { target: document.getElementById('app')!, intro: false });
