class DomainRecords extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: 'open' });
    this.recordsUrl = this.getAttribute('src') || '/records/domains.json';
  }

  async connectedCallback() {
    await this.fetchAndRender();
  }

  async fetchAndRender() {
    try {
      const response = await fetch(this.recordsUrl);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const records = await response.json();
      this.render(records);
    } catch (error) {
      this.renderError(error);
    }
  }

  groupRecordsByOwnerAndGroup(records) {
    const grouped = {};

    records.forEach(record => {
      if (!grouped[record.Owner]) {
        grouped[record.Owner] = {};
      }

      if (!grouped[record.Owner][record.GroupID]) {
        grouped[record.Owner][record.GroupID] = {
          type: record.Type,
          hostnames: []
        };
      }

      grouped[record.Owner][record.GroupID].hostnames.push(record.Hostname);
    });

    return grouped;
  }

  render(records) {
    const grouped = this.groupRecordsByOwnerAndGroup(records);

    let html = `
      <style>
        :host {
          display: block;
          font-family: inherit;
        }
        .owner {
          margin-bottom: 1em;
        }
        .owner-link {
          font-weight: bold;
          color: inherit;
        }
        .group {
          margin-left: 1.5em;
          margin-top: 0.5em;
        }
        .group-type {
          font-style: italic;
        }
        .hostnames {
          margin-left: 1em;
        }
        .hostname {
          display: inline-block;
          margin-right: 1em;
        }
        .error {
          color: #d32f2f;
          padding: 1em;
          border: 1px solid #ffcdd2;
          background-color: #ffebee;
          border-radius: 4px;
        }
      </style>
    `;

    if (Object.keys(grouped).length === 0) {
      html += '<p>No domain records found.</p>';
    } else {
      html += '<ul>';

      for (const [owner, groups] of Object.entries(grouped)) {
        html += `
          <li class="owner">
            <a href="${owner}" class="owner-link">${owner}</a>
            <ul>
        `;

        for (const [groupId, group] of Object.entries(groups)) {
          html += `
            <li class="group">
              <span class="group-type">Type: ${group.type}</span>
              <div class="hostnames">
                ${group.hostnames.map(hostname =>
                  `<span class="hostname">${hostname}</span>`
                ).join('')}
              </div>
            </li>
          `;
        }

        html += '</ul></li>';
      }

      html += '</ul>';
    }

    this.shadowRoot.innerHTML = html;
  }

  renderError(error) {
    this.shadowRoot.innerHTML = `
      <style>
        .error {
          color: #d32f2f;
          padding: 1em;
          border: 1px solid #ffcdd2;
          background-color: #ffebee;
          border-radius: 4px;
        }
      </style>
      <div class="error">
        Error loading domain records: ${error.message}
      </div>
    `;
  }
}

customElements.define('domain-records', DomainRecords);