var loopsVariables = {};
var onMount = [];

class dynamicVariable {
    constructor(name, value) {
        this.name = name;
        this.value = value;
        onMount.push(this.updateDisplay.bind(this));
    }

    set(value) {
        if (this.value === value) {
            return;
        }
        this.value = value;

        this.updateDisplay();
    }

    updateDisplay() {
        this.resetIfs();

        var elements = document.querySelectorAll(`[update~="${this.name}"]`);

        let updated = 0;
        while (updated < elements.length) {
            const element = elements[updated];
            updated++;

            if (checkIfParentHidden(element)) {
                if (element.tagName === "IF-ELEMENT") {
                    this.updateIf(element, null);
                } else if (element.tagName === "EACH-ELEMENT") {
                    this.updateLoop(element);
                    elements = document.querySelectorAll(
                        `[update~="${this.name}"]`
                    );
                } else if (element.tagName === "JS-ELEMENT") {
                    this.updateJS(element);
                }
            }
        }

        this.resetLoopsVariables();
        this.updateDynamicAttributes();
    }

    resetIfs() {
        const elements = document.querySelectorAll(
            `if-element[update~="${this.name}"]:not([hidden]), elif-element[update~="${this.name}"]:not([hidden]), else-element[update~="${this.name}"]:not([hidden])`
        );
        Array.from(elements).forEach((element) => {
            element.hidden = true;
        });
    }

    updateIf(element, from) {
        if (from) {
            const nextElement = from.nextElementSibling;
            if (nextElement) {
                const type = nextElement.tagName;
                if (type === "ELIF-ELEMENT") {
                    const condition = nextElement.attributes.js.value;
                    if (eval(condition)) {
                        nextElement.hidden = false;
                    } else {
                        nextElement.hidden = true;
                        this.updateIf(null, nextElement);
                    }
                } else if (type === "ELSE-ELEMENT") {
                    nextElement.hidden = false;
                }
            }
            return;
        } else if (element) {
            try {
                const condition = element.attributes.js.value;
                if (eval(condition)) {
                    element.hidden = false;
                } else {
                    element.hidden = true;
                    this.updateIf(null, element);
                }
            } catch (error) {}
        }
    }

    updateJS(element) {
        if ((element.attributes.html ?? { value: "false" }).value == "true") {
            element.innerHTML = eval(element.attributes.js.value);
        } else element.textContent = eval(element.attributes.js.value);
    }

    updateLoop(element) {
        try {
            let uuid = uuidV4();
            element.setAttribute("uuid", uuid);

            while (element.children.length > 1) {
                element.removeChild(element.lastChild);
            }

            const loopList = eval(element.attributes.js.value);
            loopsVariables[uuid] = loopList;

            const loopElement = element.firstElementChild;
            if (loopElement) {
                for (let i = 0; i < loopList.length; i++) {
                    const clonedElement = loopElement.cloneNode(true);
                    clonedElement.hidden = false;
                    element.appendChild(clonedElement);

                    const lastElement = element.lastElementChild;
                    const allDescendants = lastElement.querySelectorAll("*");
                    for (let k = 0; k < allDescendants.length; k++) {
                        const descendant = allDescendants[k];
                        if (Object.keys(descendant.dataset).length) {
                            for (const key in descendant.dataset) {
                                if (key.startsWith("dynamic")) {
                                    descendant.dataset[
                                        key
                                    ] = `(function() { const ${element.attributes.element.value} = loopsVariables['${uuid}'][${i}]; return ${descendant.dataset[key]}})()`;
                                }
                            }
                        }

                        if (descendant.attributes.js) {
                            const jsValue = descendant.attributes.js.value;
                            descendant.setAttribute(
                                "js",
                                `(function() { const ${element.attributes.element.value} = loopsVariables['${uuid}'][${i}]; return ${jsValue}})()`
                            );
                        }
                    }
                }
            } else {
                console.warn("No child element found to clone in:", element);
            }
        } catch (error) {}
    }

    updateDynamicAttributes() {
        const toUpdate = Array.from(
            document.querySelectorAll(`[data-update~="${this.name}"]`)
        ).filter((element) => checkIfParentHidden(element));

        toUpdate.forEach((element) => {
            if (Object.keys(element.dataset).length) {
                for (const key in element.dataset) {
                    if (key.startsWith("dynamicJs")) {
                        const attribute = key
                            .split("dynamicJs")[1]
                            .toLowerCase();
                        const attributeValue = element.dataset[key];
                        element.setAttribute(attribute, attributeValue);
                    } else if (key.startsWith("dynamic")) {
                        const attribute = key.split("dynamic")[1].toLowerCase();
                        const attributeValue = eval(element.dataset[key]);
                        element.setAttribute(attribute, attributeValue);
                    }
                }
            }
        });
    }

    resetLoopsVariables() {
        Object.keys(loopsVariables).forEach((uuid) => {
            if (!document.querySelector(`[uuid="${uuid}"]`)) {
                delete loopsVariables[uuid];
            }
        });
    }
}

function checkIfParentHidden(element) {
    let parent = element.parentElement;
    while (parent) {
        if (parent.hidden) {
            return false; // Exclude elements with hidden parents
        }
        parent = parent.parentElement;
    }
    return true; // Include elements with no hidden parents
}

function uuidV4() {
    return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(
        /[xy]/g,
        function (c) {
            var r = (Math.random() * 16) | 0,
                v = c === "x" ? r : (r & 0x3) | 0x8;
            return v.toString(16);
        }
    );
}

class jsElement extends HTMLElement {
    constructor() {
        super();
    }

    static get observedAttributes() {
        return ["js", "update", "html"];
    }
}

customElements.define("js-element", jsElement);

class ifElement extends HTMLElement {
    constructor() {
        super();
    }

    static get observedAttributes() {
        return ["js", "update"];
    }
}

customElements.define("if-element", ifElement);

class elifElement extends HTMLElement {
    constructor() {
        super();
    }

    static get observedAttributes() {
        return ["js", "update"];
    }
}

customElements.define("elif-element", elifElement);

class elseElement extends HTMLElement {
    constructor() {
        super();
    }

    static get observedAttributes() {
        return ["update"];
    }
}

customElements.define("else-element", elseElement);

class eachElement extends HTMLElement {
    constructor() {
        super();
    }

    static get observedAttributes() {
        return ["js", "update", "uuid", "element"];
    }
}

customElements.define("each-element", eachElement);

window.addEventListener("load", () => {
    console.log("load");
    const typeElements = document.querySelectorAll(
        "if-element, elif-element, else-element"
    );
    typeElements.forEach((element) => {
        element.hidden = true;
    });
    const eachElements = document.querySelectorAll("each-element");
    eachElements.forEach((element) => {
        let span = document.createElement("span");
        while (element.firstChild) {
            span.appendChild(element.firstChild);
        }
        span.hidden = true;
        element.appendChild(span);
    });
    setTimeout(() => {
        onMount.forEach((element) => {
            element();
        });
    }, 0);
});
