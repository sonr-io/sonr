import { defineConfig } from "blume";

// GitHub Pages exposes no deploy URL to the build, so the workflow forwards the
// `actions/configure-pages` outputs through these variables:
//   BLUME_SITE = the bare origin  (e.g. https://sonr-io.github.io)
//   BLUME_BASE = the subpath      (e.g. /sonr, or "/" on a custom domain)
// Both are unset locally, where Blume falls back to the dev-server origin.
const site = process.env.BLUME_SITE?.replace(/\/+$/, "") || undefined;
const rawBase = process.env.BLUME_BASE?.replace(/\/+$/, "") ?? "";
// A base of "" or "/" means the site is served at the root — omit the field
// entirely rather than mounting everything under an empty segment.
const base = rawBase && rawBase !== "/" ? rawBase : undefined;

export default defineConfig({
  title: "Sonr",
  description:
    "Developer documentation for the Sonr blockchain — decentralized identity, MPC-backed vaults, UCAN authorization, and the snrd node.",

  // Monochrome wordmark: `currentColor` fills adapt to light/dark, and `text: ""`
  // keeps the brand name from rendering twice beside the mark.
  logo: { image: { light: "/logo.svg", alt: "Sonr" }, text: "" },

  theme: {
    accent: "#17c2FF",
    action: "#279cff",
    mode: "system",
  },

  github: { owner: "sonr-io", repo: "sonr" },

  navigation: {
    // The two bulky sections get their own header tabs. A tab scopes the
    // sidebar to its own folder by URL prefix, keeping Reference's ~120 pages
    // and the 77 release notes out of the main tree.
    tabs: [
      { label: "Reference", path: "/reference", icon: "code" },
      // /changelog is a generated timeline, not a content page, so a bare tab
      // path resolves to the newest entry. `href` lands on the index instead.
      { label: "Changelog", path: "/changelog", href: "/changelog", icon: "hourglass" },
    ],
    // Collapsible groups, so the main sidebar reads as categories rather than
    // one undifferentiated list.
    sidebar: { display: "group" },
    featured: [{ label: "sonr.io", href: "https://sonr.io", icon: "globe" }],
  },

  analytics: {
    scripts: [
      {
        src: "https://www.googletagmanager.com/gtag/js?id=G-FY0WZBSGH8",
        strategy: "async",
      },
      {
        content: [
          "window.dataLayer = window.dataLayer || [];",
          "function gtag(){dataLayer.push(arguments);}",
          'gtag("js", new Date());',
          'gtag("config", "G-FY0WZBSGH8");',
        ].join("\n"),
      },
    ],
  },

  // The guide categories were flattened out of /guides/* onto the site root,
  // and the Cryptography section was dropped in favour of the sonr-io/crypto
  // docs site. Keep the old URLs resolving.
  redirects: [
    {
      "from": "/guides/development",
      "to": "/development"
    },
    {
      "from": "/guides/configure-local-node",
      "to": "/configure-local-node"
    },
    {
      "from": "/guides/register-records",
      "to": "/register-records"
    },
    {
      "from": "/guides/authorize-clients",
      "to": "/authorize-clients"
    },
    {
      "from": "/guides/broadcast-transactions",
      "to": "/broadcast-transactions"
    },
    {
      "from": "/guides/issue-payments",
      "to": "/issue-payments"
    },
    {
      "from": "/guides/onboarding",
      "to": "/onboarding"
    },
    {
      "from": "/guides/deployment",
      "to": "/deployment"
    },
    {
      "from": "/guides/vrf-key-management",
      "to": "/vrf-key-management"
    },
    {
      "from": "/guides/vrf-migration",
      "to": "/vrf-migration"
    },
    {
      "from": "/guides/architecture/delegated-proof-of-stake",
      "to": "/architecture/delegated-proof-of-stake"
    },
    {
      "from": "/guides/architecture/inter-blockchain-communication",
      "to": "/architecture/inter-blockchain-communication"
    },
    {
      "from": "/guides/architecture/passkey-web-authentication",
      "to": "/architecture/passkey-web-authentication"
    },
    {
      "from": "/guides/economics/distribution",
      "to": "/economics/distribution"
    },
    {
      "from": "/guides/economics/governance",
      "to": "/economics/governance"
    },
    {
      "from": "/guides/economics/rewards",
      "to": "/economics/rewards"
    },
    {
      "from": "/guides/economics/staking",
      "to": "/economics/staking"
    },
    {
      "from": "/guides/economics/utility",
      "to": "/economics/utility"
    },
    {
      "from": "/guides/economics/values",
      "to": "/economics/values"
    },
    {
      "from": "/guides/security/audit-report",
      "to": "/security/audit-report"
    },
    {
      "from": "/guides/security/compliance",
      "to": "/security/compliance"
    },
    {
      "from": "/guides/security/cryptography",
      "to": "/security/cryptography"
    },
    {
      "from": "/guides/security/vulnerabilities",
      "to": "/security/vulnerabilities"
    },
    {
      "from": "/guides",
      "to": "/development"
    },
    {
      "from": "/cryptography",
      "to": "/security/cryptography"
    },
    {
      "from": "/cryptography/primitives",
      "to": "/security/cryptography"
    },
    {
      "from": "/cryptography/keys-and-identifiers",
      "to": "/security/cryptography"
    },
    {
      "from": "/cryptography/mpc-enclave",
      "to": "/security/cryptography"
    },
    {
      "from": "/cryptography/ucan",
      "to": "/security/cryptography"
    },
    {
      "from": "/cryptography/webauthn",
      "to": "/security/cryptography"
    },
    {
      "from": "/cryptography/did-documents",
      "to": "/security/cryptography"
    },
    {
      "from": "/cryptography/vault-encryption",
      "to": "/security/cryptography"
    },
    {
      "from": "/cryptography/vrf",
      "to": "/security/cryptography"
    },
    {
      "from": "/cryptography/transaction-authorization",
      "to": "/security/cryptography"
    },
    {
      "from": "/cryptography/plugin-integrity",
      "to": "/security/cryptography"
    },
    {
      "from": "/cryptography/implementation-status",
      "to": "/security/cryptography"
    }
  ],

  lastModified: true,

  ...(site ? { deployment: { site, ...(base ? { base } : {}) } } : {}),
});
