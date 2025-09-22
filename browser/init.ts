import { elementFetch } from './fetch'

const initFetch = "init-fetch";

async function initFunc(el: Element) {
    if (!el.hasAttribute(initFetch)) return;
    el.dispatchEvent(new CustomEvent(initFetch, { bubbles: true }));
    var endpoint = el.getAttribute(initFetch)
    el.removeAttribute(initFetch);
    elementFetch(el, endpoint)
}

new MutationObserver(ms => {
    for (const m of ms) {
        if (m.target instanceof HTMLElement) initFunc(m.target);
    }
}).observe(document.body, {
    childList: true,
    subtree: true,
    attributes: true,
    attributeFilter: [initFetch]
});

document.querySelectorAll(`[${initFetch}]`).forEach(el => {
    if (el instanceof HTMLElement) initFunc(el);
});