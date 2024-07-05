// Usage: <password-input name="password" placeholder="Enter your password"/>
class PasswordInput extends HTMLElement {
    private show: boolean;

    constructor() {
        super();
        this.show = false;
    }

    // TODO: Add interface with the 4 callbacks that custom elements can have.
    connectedCallback() {
        const name = this.getAttribute("name");
        const placeholder = this.getAttribute("placeholder")
        this.innerHTML = `
            <label for="${name}" class="input input-bordered flex items-center gap-2">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="w-4 h-4 opacity-70"><path fill-rule="evenodd" d="M14 6a4 4 0 0 1-4.899 3.899l-1.955 1.955a.5.5 0 0 1-.353.146H5v1.5a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1-.5-.5v-2.293a.5.5 0 0 1 .146-.353l3.955-3.955A4 4 0 1 1 14 6Zm-4-2a.75.75 0 0 0 0 1.5.5.5 0 0 1 .5.5.75.75 0 0 0 1.5 0 2 2 0 0 0-2-2Z" clip-rule="evenodd" /></svg>
              <div class="relative w-full"/>
                  <input type="password" placeholder="${placeholder}" name="${name}" class="grow" required/>
                  <div data-role="show" class="absolute inset-y-0 right-0 flex cursor-pointer text-gray-700">Show</div>
              </div>
            </label>
        `
        this.querySelector('div[data-role="show"]').addEventListener('click', () => {
            this.show = !this.show;
            this.querySelector('input').setAttribute('type', this.show ? "text" : "password");
        });
    }
}

customElements.define('password-input', PasswordInput);
