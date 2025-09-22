import './browser'

function applyJson(updates: Update[]) {
    for (let index = 0; index < updates.length; index++) {
        const update = updates[index];
        const targets = document.querySelectorAll(update.selector);
        if (targets == null || targets.length == 0) {
            console.error(`no target found for selector ${update.selector}`)
            return;
        }
        for (let index = 0; index < targets.length; index++) {
            const target = targets[index];
            for (const [key, value] of Object.entries(update.text ?? {})) {
                target[key] = value;
            }
            for (const [key, value] of Object.entries(update.integer ?? {})) {
                target[key] = value;
            }
            for (const [key, value] of Object.entries(update.boolean ?? {})) {
                target[key] = value;
            }
            for (const [key, value] of Object.entries(update.attribute ?? {})) {
                target.setAttribute(key, value);
            }
            for (const [key, value] of Object.entries(update.style ?? {})) {
                (target as HTMLElement).style.setProperty(key, value);
            }
            if (update.remove ?? false) {
                target.remove();
            }
            // maybe?
            // https://developer.mozilla.org/en-US/docs/Web/API/Element/animate
            // https://developer.mozilla.org/en-US/docs/Web/API/EventTarget/dispatchEvent
            // https://developer.mozilla.org/en-US/docs/Web/API/Node/normalize
            // https://developer.mozilla.org/en-US/docs/Web/API/Element/releasePointerCapture
            // https://developer.mozilla.org/en-US/docs/Web/API/Element/setPointerCapture
            // https://developer.mozilla.org/en-US/docs/Web/API/Element/requestFullscreen
            // https://developer.mozilla.org/en-US/docs/Web/API/Element/requestPointerLock
            // https://developer.mozilla.org/en-US/docs/Web/API/Element/scrollIntoView
        }
    }
};

function applyHtml(
    selector: string,
    event: string,
    template: HTMLTemplateElement
) {
    let elements = template.content.children;
    for (let index = 0; index < elements.length; index++) {
        const element = elements[index];
        if (element.tagName == 'SCRIPT') {
            const newScript = document.createElement("script");
            newScript.textContent = element.textContent;
            elements[index].replaceWith(newScript);
        }
    }

    if (event == "") { // no event -> replace by ID
        for (var child of Array.from(elements)) {
            const target = document.getElementById((child as Element).id)
            if (target == null) {
                console.error(`no target found for id ${(child as Element).id}`)
                continue
            }
            target.replaceWith(child)
        }
        return
    }

    const targets = document.querySelectorAll(selector)
    if (targets == null || targets.length == 0) {
        console.error(`no target found for selector ${selector}`)
        return;
    }
    for (let index = 0; index < targets.length; index++) {
        const target = targets[index];
        if (event == "replace") {
            target.replaceWith(...elements)
        } else if (event == "prepend") {
            target.prepend(...elements)
        } else if (event == "append") {
            target.append(...elements)
        } else if (event == "after") {
            target.after(...elements)
        } else if (event == "before") {
            target.before(...elements)
        }
    }
};


export async function handleResponse(
    response: Response
) {
    if (response.status >= 400) {
        console.error(`status ${response.status}, url ${response.url}`);
        return;
    }

    var localApplyHtml = applyHtml
    if ('startViewTransition' in document) {
        localApplyHtml = (selector: string, event: string, template: HTMLTemplateElement) => {
            document.startViewTransition(() => {
                applyHtml(selector, event, template)
            })
        }
    }

    var localApplyJson = applyJson
    if ('startViewTransition' in document) {
        localApplyJson = (updates: Update[]) => {
            document.startViewTransition(() => {
                applyJson(updates)
            })
        }
    }

    if (response.headers.get('Content-Type')?.includes('text/html')) {
        const template = document.createElement('template')
        template.innerHTML = await response.text();
        localApplyHtml(
            response.headers.get('X-Id') || '',
            response.headers.get('X-Event') || '',
            template,
        )
        return
    }

    if (response.headers.get('Content-Type')?.includes('application/json')) {
        localApplyJson(JSON.parse(await response.text()) as Update[]);
        return
    }

    if (!response.headers.get('Content-Type')?.includes('text/event-stream')) {
        return
    }


    const reader = response.body!.getReader()
    const decoder = new TextDecoder('utf-8')

    let buffer = ''
    let partialEventLines: string[] = []

    while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })

        const lines = buffer.split('\n')
        buffer = lines.pop() || ''
        for (let line of lines) {
            if (line.trim() !== '') {
                partialEventLines.push(line)
            } else {
                let event = ''
                let id = ''

                for (let i = 0; i < partialEventLines.length; i++) {
                    let eventLine = partialEventLines[i]
                    if (eventLine.startsWith('event:')) {
                        event = eventLine.slice(6).trim()
                        partialEventLines[i] = ''
                    } else if (eventLine.startsWith('id:')) {
                        id = eventLine.slice(3).trim()
                        partialEventLines[i] = ''
                    } else if (eventLine.startsWith('data:')) {
                        partialEventLines[i] = eventLine.slice(5).trim()
                    } else {
                        partialEventLines[i] = ''
                    }
                }

                if (event == 'dom') {
                    localApplyJson(JSON.parse(partialEventLines.join('\n')) as Update[]);
                } else {
                    const template = document.createElement('template')
                    template.innerHTML = partialEventLines.join('\n')
                    localApplyHtml(id, event, template)
                }
                partialEventLines = []
            }
        }
    }
};

export function extractInputs(el: Element): string {
    const inputs = el.querySelectorAll('input, select, textarea')
    const formData = new FormData();
    for (const input of inputs) {
        if (
            (
                input instanceof HTMLInputElement ||
                input instanceof HTMLTextAreaElement ||
                input instanceof HTMLSelectElement
            ) &&
            input.name
        ) {
            if (input instanceof HTMLInputElement && input.type === 'checkbox') {
                formData.append(input.name, input.checked ? '1' : '0')
            } else {
                formData.append(input.name, input.value)
            }
        }
    }
    return JSON.stringify(Object.fromEntries(formData))
};

export async function elementFetch(el: Element, endpoint: string | null) {
    if (endpoint == null || endpoint == '') {
        console.warn(`endpoint not specified for element`)
        console.log(el as Object)
        return;
    }
    if (endpoint == "window.location.pathname") {
        endpoint = window.location.pathname
    }
    if (endpoint.startsWith("eval")) {
        endpoint = window.eval(endpoint)
    }
    var method = el.getAttribute('method') || 'GET'
    var success = false
    for (let index = 0; index <= Number(el.getAttribute('retry')); index++) {
        let init: RequestInit = {
            method: method,
        }
        init.body = method == 'GET' ? null : extractInputs(el)

        await fetch(endpoint!, init).then(handleResponse)
            .then(() => {
                success = true
            })
            .catch((e) => {
                console.warn(`retry: ${index}, method: ${method}, url ${endpoint}, body: ${init.body}, error: ${e}`)
                success = false
            })
        if (success) {
            break
        }
        await new Promise(resolve => setTimeout(resolve, 1000));
    }
};