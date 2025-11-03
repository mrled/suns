class DomainRecords extends HTMLElement {
  constructor() {
    super();
    this.recordsUrl = this.getAttribute('src') || '/records/domains.json';

    // Mapping of short character group types to human readable types
    // Based on symgroup.go in symval/internal/symgroup/
    this.typeCodeToName = {
      'a': 'Palindrome',
      'b': 'Flip 180',
      'c': 'Double Flip 180',
      'd': 'Mirror Text',
      'e': 'Mirror Names',
      'f': 'Antonym Names'
    };
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

  getHumanReadableType(typeCode) {
    return this.typeCodeToName[typeCode] || typeCode;
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
        domain-records {
          display: block;
          font-family: inherit;
        }
      </style>
    `;

    if (Object.keys(grouped).length === 0) {
      html += '<p>No domain records found.</p>';
    } else {
      html += '<ul>';

      for (const [owner, groups] of Object.entries(grouped)) {
        html += `<li class="owner"><a href="${owner}">${owner}</a><ul>`;
        for (const [groupId, group] of Object.entries(groups)) {
          const humanReadableType = this.getHumanReadableType(group.type);
          const domainList = group.hostnames.map(h => `<code>${h}</code>`).join(', ');
          html += `<li><span>${humanReadableType}</span>: ${domainList}</li>`;
        }

        html += '</ul></li>';
      }

      html += '</ul>';
    }

    this.innerHTML = html;
  }

  renderError(error) {
    this.innerHTML = `
      <style>
        domain-records .error {
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