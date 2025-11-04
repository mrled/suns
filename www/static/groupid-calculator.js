class GroupIDCalculator extends HTMLElement {
  constructor() {
    super();
  }

  connectedCallback() {
    this.render();
    this.attachEventHandlers();
  }

  async calculateGroupID(owner, type, hostnames) {
    // Validate inputs
    if (!owner) {
      throw new Error('Owner cannot be empty');
    }
    if (!type) {
      throw new Error('Type cannot be empty');
    }
    if (!hostnames || hostnames.length === 0) {
      throw new Error('At least one hostname is required');
    }

    // Sort hostnames for consistent hashing
    const sorted = [...hostnames].sort();

    // Hash the owner using SHA-256
    const ownerBytes = new TextEncoder().encode(owner);
    const ownerHash = await crypto.subtle.digest('SHA-256', ownerBytes);
    const ownerEncoded = btoa(String.fromCharCode(...new Uint8Array(ownerHash)));

    // Concatenate all sorted hostnames
    const hostnamesString = sorted.join('');

    // Hash the hostnames using SHA-256
    const hostnamesBytes = new TextEncoder().encode(hostnamesString);
    const hostnamesHash = await crypto.subtle.digest('SHA-256', hostnamesBytes);
    const hostnamesEncoded = btoa(String.fromCharCode(...new Uint8Array(hostnamesHash)));

    // Format: v1:type:base64(sha256(owner)):base64(sha256(sort(hostnames)))
    return `v1:${type}:${ownerEncoded}:${hostnamesEncoded}`;
  }

  attachEventHandlers() {
    const form = this.querySelector('form');
    if (!form) return;

    form.addEventListener('submit', async (e) => {
      e.preventDefault();

      const owner = this.querySelector('#owner').value.trim();
      const type = this.querySelector('#type').value.trim();
      const hostnamesText = this.querySelector('#hostnames').value.trim();

      // Parse hostnames (space-separated)
      const hostnames = hostnamesText
        .split(/\s+/)
        .map(h => h.trim())
        .filter(h => h);

      const resultField = this.querySelector('#result');

      try {
        const groupID = await this.calculateGroupID(owner, type, hostnames);
        resultField.value = groupID;
      } catch (error) {
        resultField.value = `Error: ${error.message}`;
      }
    });
  }

  render() {
    this.innerHTML = `
      <style>
        groupid-calculator {
          display: block;
          font-family: inherit;
        }
        groupid-calculator form {
          display: flex;
          flex-direction: column;
          gap: 1rem;
          max-width: 600px;
          margin: 1em;
        }
        groupid-calculator label {
          display: flex;
          flex-direction: column;
          gap: 0.25rem;
        }
        groupid-calculator input,
        groupid-calculator select {
          font-family: inherit;
          font-size: inherit;
          padding: 0.5rem;
          border: 1px solid var(--body-fg-color);
          color: inherit;
          background-color: inherit;
        }
        groupid-calculator .buttons {
          display: flex;
          gap: 1rem;
        }
        groupid-calculator button {
          padding: 0.5rem 1rem;
          font-family: inherit;
          font-size: inherit;
          cursor: pointer;
          color: inherit;
          background-color: inherit;
          border: 1px solid var(--body-fg-color);
        }
      </style>

      <h2>Group ID Calculator</h2>

      <form>
        <label>
          Owner:
          <input type="text" id="owner" placeholder="https://example.com" required>
        </label>

        <label>
          Type:
          <select id="type" required>
            <option value="">Select a type...</option>
            <option value="a">a - Palindrome</option>
            <option value="b">b - Flip 180</option>
            <option value="c">c - Double Flip 180</option>
            <option value="d">d - Mirror Text</option>
            <option value="e">e - Mirror Names</option>
            <option value="f">f - Antonym Names</option>
          </select>
        </label>

        <label>
          Hostnames (space separated):
          <input type="text" id="hostnames" placeholder="host1.example.com host2.example.com" required>
        </label>

        <div class="buttons">
          <button type="submit">Calculate Group ID</button>
        </div>

        <label>
          Result:
          <input type="text" id="result" readonly>
        </label>
      </form>
    `;
  }
}

customElements.define('groupid-calculator', GroupIDCalculator);