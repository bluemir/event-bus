import $ from '../lib/web-components/minilib.module.js';
import {html, render} from 'https://unpkg.com/lit-html?module';


var template = (text) => html`
<style>
	:host {
	}
</style>
<h1>Hello World</h1>
<p>${text}</p>
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
		var ws = new WebSocket($.util.wsURL(`/v1/stream?token=${$.util.base64.encode("test:")}`));
	}
}
customElements.define("app-main", AppMain);
