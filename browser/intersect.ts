import { elementFetch } from './fetch'

const intersectFetch = "intersect-fetch";
const intersectObserved = "intersect-observed";

async function intersectFunc(el: Element) {
    if (!el.hasAttribute(intersectFetch)) return;
    el.dispatchEvent(new CustomEvent(intersectFetch, { bubbles: true }));
    var endpoint = el.getAttribute(intersectFetch)
    el.removeAttribute(intersectFetch);
    elementFetch(el, endpoint)
}

let intObs = new IntersectionObserver((entries) => {
    for (const entry of entries) {
        if (entry.isIntersecting) {
            intersectFunc(entry.target)
            intObs.unobserve(entry.target);
        }
    }
},)

new MutationObserver(ms => {
    for (const m of ms) {
        if (m.target instanceof HTMLElement) {
            if (m.target.getAttribute(intersectObserved) || !m.target.getAttribute(intersectFetch)) {
                return
            }
            intObs.observe(m.target);
            m.target.setAttribute(intersectObserved, "true");
        };
    }
}).observe(document.body, {
    childList: true,
    subtree: true,
    attributes: true,
    attributeFilter: [intersectFetch]
});

document.querySelectorAll(`[${intersectFetch}]`).forEach((el) => {
    if (el instanceof HTMLElement) {
        if (el.getAttribute(intersectObserved) || !el.getAttribute(intersectFetch)) {
            return
        }
        intObs.observe(el);
        el.setAttribute(intersectObserved, "true");
    }
});