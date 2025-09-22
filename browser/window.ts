window.addEventListener('keydown', (event) => {
    if (event.key == 'Escape') {
        return (document.activeElement as any)?.blur();
    }
    if (['Shift', 'Alt', 'Meta', 'Enter', 'Escape'].includes(event.key)) { return; }
    if (event.ctrlKey || event.metaKey) { return; }
    if (event.target instanceof HTMLInputElement && event.target.type === "number") {
        ['e', 'E', '+', '-', '.', ','].includes(event.key) && event.preventDefault()
        if (event.target.value !== '') {
            const cleaned = String(Number(event.target.value));
            if (event.target.value !== cleaned) {
                event.target.value = cleaned;
            }
        } else {
            event.target.value = "0"
        }
    }
    let target = document.querySelectorAll(`.active [key-down='${event.key}'], #navigation [key-down='${event.key}']`);
    if (target.length > 0) {
        event.preventDefault();
        (target[0] as any).scrollIntoViewIfNeeded();
        (target[0] as HTMLElement).focus();
        (target[0] as HTMLElement).click();
    }
});

function setActivePage() {
    var hash = window.location.hash
    if (hash == null || hash == '') {
        console.warn('empty hash');
        return
    }
    document.querySelectorAll(`section.page.active`).forEach((el) => {
        el.classList.remove('active')
    })
    document.querySelectorAll(`section.page[name="${hash.slice(1)}"]`).forEach((el) => {
        el.classList.add('active')
    })
}

setActivePage()
window.addEventListener('hashchange', setActivePage)

window.addEventListener('click', (event) => {
    var el = event.target as Element
    var scrollY = el.getAttribute("click-scroll-y")
    if (!scrollY) {
        return;
    }
    window.scrollBy(0, scrollY as any);
})