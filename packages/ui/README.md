# @sonr.io/ui

Centralized UI component library for the Sonr ecosystem using shadcn/ui.

## 🎨 Overview

This package provides a fully centralized component library following shadcn/ui patterns. All UI components, styles, and utilities are managed exclusively in this package and imported by web applications. No components are duplicated in individual apps.

## 📦 Installation

The package is already configured in the monorepo. Web applications import components directly:

```tsx
import { Button, Alert, Input } from "@sonr.io/ui";
import { cn } from "@sonr.io/ui";
```

## 🚀 Adding New Components

**All components are managed centrally in this package.** To add new shadcn components:

```bash
# Navigate to the UI package (REQUIRED)
cd packages/ui

# Add a new component using shadcn CLI
npx shadcn@latest add dialog

# The component will be added to src/components/ui/
```

⚠️ **Important**: Never run `npx shadcn add` from web application directories. All components must be added here.

## 📁 Structure

```
packages/ui/
├── src/
│   ├── components/
│   │   ├── ui/          # shadcn components (managed by CLI)
│   │   │   ├── button.tsx
│   │   │   ├── input.tsx
│   │   │   └── alert.tsx
│   │   └── index.ts     # Main exports
│   ├── lib/
│   │   └── utils.ts     # cn utility function
│   └── styles/
│       └── globals.css  # Theme and CSS variables
├── components.json      # shadcn CLI configuration
├── tailwind.config.js   # Tailwind configuration
└── README.md           # This file
```

## 🎨 Theming

The primary brand color is `#17c2ff` (cyan). All theme variables are defined in `src/styles/globals.css`:

- **Light mode**: CSS variables under `:root`
- **Dark mode**: CSS variables under `.dark`
- **Primary color**: HSL(195, 100%, 54%)

## 💻 Usage in Applications

### 1. Import Global Styles

In your app's `globals.css`:

```css
@import "@sonr.io/ui/styles/globals.css";

/* App-specific styles below if needed */
```

### 2. Import Components

```tsx
import { Button, Input, Alert, AlertTitle, AlertDescription } from "@sonr.io/ui";
import { cn } from "@sonr.io/ui";

export function MyComponent() {
  return (
    <div className="space-y-4">
      <Input placeholder="Enter text..." />
      
      <Button variant="default" size="lg">
        Click me
      </Button>
      
      <Alert>
        <AlertTitle>Success!</AlertTitle>
        <AlertDescription>Your action was completed.</AlertDescription>
      </Alert>
    </div>
  );
}
```

## 🧩 Available Components

### Core Components (shadcn/ui)

- **Button**: Multiple variants (default, destructive, outline, secondary, ghost, link)
- **Input**: Styled form input with full accessibility
- **Alert**: Alert messages with title and description support
- **Card**: Container with header, content, and footer sections

### Utility Functions

- **cn()**: Class name utility for merging Tailwind classes

## 🛠️ Development

```bash
# Install dependencies
pnpm install

# Lint the package
pnpm lint

# Type check
pnpm exec tsc --noEmit

# Add new shadcn component
npx shadcn@latest add [component-name]
```

## 📋 Centralized Workflow

This monorepo follows a **fully centralized** UI component strategy:

1. ✅ **Single Source of Truth**: All UI components live only in `packages/ui`
2. ✅ **No Duplication**: Web apps do not have their own UI components
3. ✅ **Consistent Theming**: All apps share the exact same theme
4. ✅ **Simplified Maintenance**: Update once, affects all apps
5. ✅ **shadcn CLI Management**: Run all shadcn commands from `packages/ui` only

## ⚠️ Important Guidelines

- **Never** create components.json in web applications
- **Never** run `npx shadcn add` from app directories  
- **Always** add new components from the `packages/ui` directory
- **Always** export new components from `src/components/index.ts`

## 🔄 Migration from Distributed to Centralized

If migrating from a distributed component structure:

1. Remove any `components.json` files from web apps
2. Delete any shadcn components from app directories
3. Update imports to use `@sonr.io/ui`
4. Import global styles from the UI package

## 📝 Component Addition Checklist

When adding a new shadcn component:

- [ ] Navigate to `packages/ui` directory
- [ ] Run `npx shadcn@latest add [component]`
- [ ] Export component from `src/components/index.ts`
- [ ] Test import in a web application
- [ ] Update this README with the new component

## 🚀 Future Enhancements

- [ ] Storybook for component documentation
- [ ] Unit tests for all components
- [ ] Additional shadcn components as needed
- [ ] Custom Sonr-specific components