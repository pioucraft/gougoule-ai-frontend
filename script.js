var loopsVariables = {};

class dynamicVariable {
    constructor(name, value) {
        this.name = name;
        this.value = value;

        setTimeout(() => {
            this.updateDisplay();
        }, 0);
    }

    set(value) {
        this.value = value;

        this.updateDisplay();
    }

    updateDisplay() {
        this.resetIfs();

        this.updateLoops();
        this.updateIfs();
        this.updateJS();
    }

    resetIfs() {
        const elements = document.querySelectorAll(
            `if-element[update~="${this.name}"]:not([hidden]), elif-element[update~="${this.name}"]:not([hidden]), else-element[update~="${this.name}"]:not([hidden])`
        );
        Array.from(elements).forEach((element) => {
            element.hidden = true;
        });
    }

    updateIfs(from) {
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
                        this.updateIfs(nextElement);
                    }
                } else if (type === "ELSE-ELEMENT") {
                    nextElement.hidden = false;
                }
            }
            return;
        }

        const elements = document.querySelectorAll(
            `if-element[update~="${this.name}"]`
        );
        Array.from(elements).forEach((element) => {
            try {
                const condition = element.attributes.js.value;
                if (eval(condition)) {
                    element.hidden = false;
                } else {
                    element.hidden = true;
                    this.updateIfs(element);
                }
            } catch (error) {}
        });
    }

    updateJS() {
        const elements = Array.from(
            document.querySelectorAll(
                `js-element[update~="${this.name}"]:not([hidden])`
            )
        ).filter((element) => {
            let parent = element.parentElement;
            while (parent) {
                if (parent.hidden) {
                    return false; // Exclude elements with hidden parents
                }
                parent = parent.parentElement;
            }
            return true; // Include elements with no hidden parents
        });
        Array.from(elements).forEach((element) => {
            if (
                (element.attributes.html ?? { value: "false" }).value == "true"
            ) {
                element.innerHTML = eval(element.attributes.js.value);
            } else element.textContent = eval(element.attributes.js.value);
        });
    }

    updateLoops() {
        let elements = document.querySelectorAll(
            `each-element[update~="${this.name}"]`
        );
        let updated = 0;
        while (updated < elements.length) {
            const element = elements[updated];

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
                        const allDescendants =
                            lastElement.querySelectorAll("*");
                        for (let k = 0; k < allDescendants.length; k++) {
                            const descendant = allDescendants[k];
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
                    console.warn(
                        "No child element found to clone in:",
                        element
                    );
                }
            } catch (error) {}
            elements = document.querySelectorAll(
                `each-element[update~="${this.name}"]`
            );
            updated++;
        }
        Object.keys(loopsVariables).forEach((uuid) => {
            if (!document.querySelector(`[uuid="${uuid}"]`)) {
                delete loopsVariables[uuid];
            }
        });
    }
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
        return ["js"];
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

window.onload = function () {
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
};
