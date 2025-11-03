+++
title = "Society for Universal Name Symmetry"
draft = true
+++

A test page for the web component(s).

Keep this page as a draft, it is only used for development purposes.

<script src="/domain-records.js"></script>

## Default domain records for this site

The default domain records for this site are set by the `recordsDomainsPath` param in Hugo settings.
For this environment, the value for that param is `{{< param "recordsDomainsPath" >}}`.

<domain-records src='{{< param "recordsDomainsPath" >}}'></domain-records>

## Domain records for `/records-domains-example.json`

This path is part of our Hugo site and contains an example set of records.

<domain-records src="/records-domains-example.json"></domain-records>

## Domain records for `https://zq.suns.bz/records/domains.json`

This path is the production record set.
It's available on any site at this URL,
and has permissive CORS settings to allow fetching from anywhere.

<domain-records src="https://zq.suns.bz/records/domains.json"></domain-records>
