import { elementFetch } from './fetch'

const inputFetch = "input-fetch";

async function inputFunc(el: Element) {
    if (!el.hasAttribute(inputFetch)) return;
    el.addEventListener('input', (ev) => {
        elementFetch(el, el.getAttribute(inputFetch));
    })
}

new MutationObserver(ms => {
    for (const m of ms) {
        for (const a of m.addedNodes) {
            if (a instanceof HTMLElement) inputFunc(a);
        }
    }
}).observe(document.body, {
    childList: true,
    subtree: true,
    attributes: true,
    attributeFilter: [inputFetch]
});

document.querySelectorAll(`[${inputFetch}]`).forEach(el => {
    if (el instanceof HTMLElement) inputFunc(el);
});