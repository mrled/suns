class DomainRecords extends HTMLElement {
  constructor() {
    super();

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
    // Get the URL when the element is connected to the DOM
    this.recordsUrl = this.getAttribute('src') || '/records/domains.json';

    // Get priority owner attribute (single owner to show first)
    this.priorityOwner = this.getAttribute('priority-owner') || null;

    await this.fetchAndRender();
  }

  async fetchAndRender() {
    try {
      const response = await fetch(this.recordsUrl);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const records = await response.json();
      console.log(`Fetched domain records from ${this.recordsUrl}:`, records);
      this.render(records);
    } catch (error) {
      console.error(`Error fetching domain records from ${this.recordsUrl}:`, error);
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

  sortOwnersByPriority(owners) {
    if (!this.priorityOwner) {
      // No priority specified, return alphabetically sorted
      return owners.sort();
    }

    // Sort all owners alphabetically first
    const sortedOwners = [...owners].sort();

    // If priority owner exists in the list, move it to the front
    const priorityIndex = sortedOwners.indexOf(this.priorityOwner);
    if (priorityIndex > -1) {
      // Remove from current position and add to beginning
      sortedOwners.splice(priorityIndex, 1);
      sortedOwners.unshift(this.priorityOwner);
    }

    return sortedOwners;
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

      // Sort owners based on priority
      const sortedOwners = this.sortOwnersByPriority(Object.keys(grouped));

      for (const owner of sortedOwners) {
        const groups = grouped[owner];
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
          color: #ff0000ff;
        }
      </style>
      <p class="error">
        Error loading domain records: ${error.message}
      </p>
    `;
  }
}

customElements.define('domain-records', DomainRecords);