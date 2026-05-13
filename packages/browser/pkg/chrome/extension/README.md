# Chrome Extension Boilerplate

A Chrome extension boilerplate with Vite, TypeScript, and React. It also uses Mantine components for the UI.

## Specific to this project

tbd

## Prerequisites

- Node.js v22 or higher
- pnpm v10 or higher

## What's included

- Vite
- TypeScript
- React
- Mantine components
- Tailwind CSS
- Tabler icons
- Magic UI
- Shadcn UI

## Installation

```bash
# Install dependencies
pnpm install
```

## Development

```bash
# Start the dev server
pnpm run dev
```

The dev server will run at `http://localhost:5173/`

- View the popup at: `http://localhost:5173/src/popup/`

## Load in Chrome for Development

1. Run `pnpm run build` to create the `dist` folder
2. Open Chrome and navigate to `chrome://extensions/`
3. Enable "Developer mode" (toggle in top-right corner)
4. Click "Load unpacked"
5. Select the `dist` folder from this project

## Build

```bash
# Build the extension
pnpm run build
```

This creates a production-ready extension in the `dist/` folder.

## Process for releasing a new version

1. Run `pnpm version patch` to update the version in the `package.json` file
2. Run `pnpm run build:zip` to create the `zips/extension-v${pkg.version}.zip` file
3. Upload the `zips/extension-v${pkg.version}.zip` file to the Chrome Web Store

## Icon Generation

Visit here: https://alexleybourne.github.io/chrome-extension-icon-generator/

## License

ISC
