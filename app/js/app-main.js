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
	}
	render() {
		render(template(this.attr("text")), this.shadow);
	}
}
customElements.define("app-main", AppMain);
