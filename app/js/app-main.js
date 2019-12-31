import $ from '../lib/web-components/minilib.module.js';
import {html, render} from 'https://unpkg.com/lit-html?module';


var template = (text) => html`
<style>
	:host {
		color: green;
	}
</style>
<p>${text}</p>
<slot></slot>
`


class AppMain  extends $.CustomElement {
	constructor() {
		super();

		this.render();
		this.test();
	}
	render() {
		render(template(this.attr("text")), this.shadow);
	}
	test() {
		var ws = new WebSocket($.util.wsURL(`/v1/stream?token=${btoa("test:")}`));
	}
}
customElements.define("app-main", AppMain);
