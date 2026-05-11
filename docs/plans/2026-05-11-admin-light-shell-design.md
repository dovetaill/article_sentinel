# Admin Light Shell Design

**Date:** 2026-05-11

## Goal

Repair the admin frontend shell so it matches a classic Mainland government / enterprise back-office aesthetic: dark sidebar only, light workspace on the right, white header/tabs/cards/tables, and Ant Design blue as the accent. The fix must preserve the current Umi + React + Ant Design / ProComponents stack, existing routes, services, and page behavior.

## Current Code Reality

The current `web/admin` frontend is not the Vite-style structure described in the request. The active implementation uses:

- `web/admin/config/defaultSettings.ts` for ProLayout settings
- `web/admin/src/layouts/BasicLayout.tsx` for the shell and tabbed workspace
- `web/admin/src/global.less` for layout and component styling
- page-level list/detail screens under `web/admin/src/pages/**`

The dark visual spillover comes from the shell layer, not the business logic. In particular:

- ProLayout is configured with `layout: 'mix'`
- ProLayout is configured with `navTheme: 'realDark'`
- the tab strip is rendered by `PageTabs`
- page cards and content backgrounds are already intended to be light, but the shell semantics and tab/header treatment make the workspace feel partially dark and visually inconsistent

## Accepted Direction

Use a classic side-layout shell as the baseline:

- keep a dark sidebar
- move the layout from `mix` to `side`
- keep the existing route/tab workspace behavior
- restyle the header/tabs/content surfaces to light mode
- explicitly use a light Ant Design theme configuration with `defaultAlgorithm`
- keep all business pages, routes, requests, and actions unchanged

## Visual Direction

### Sidebar

- deep navy / blue-black background
- light text
- blue active item state
- dark treatment scoped to sidebar/menu only

### Workspace

- page background: light gray (`#f0f2f5` / `#f5f7f9` family)
- header, tabs, cards, filters, and tables: white surfaces
- borders: thin, cool light gray
- text: dark desaturated blue-gray for stronger enterprise readability
- motion and decoration: minimal, stable, not flashy

### Tabs / Header

- white tab rail
- compact tabs
- active tab uses blue emphasis with subtle blue-tinted background or underline
- no black outlines, empty dark frames, or oversized borders

### Page Surfaces

- statistics cards: white
- filter toolbars: white
- table wrappers: white
- table headers: `#fafafa`
- row dividers: `#f0f0f0`

## Scope

### In Scope

- ProLayout shell configuration
- runtime Ant Design theme setup
- global layout / surface CSS tokens and classes
- `PageTabs` visual contract
- shared list-page surface classes
- `/inspection/tasks` first, then equivalent list pages that share the same shell conventions
- tests for dark-sidebar/light-workspace contracts

### Out of Scope

- backend APIs
- request payload structure
- service modules under `src/services/*`
- routing semantics
- permission logic
- task creation/query/delete logic
- migration to another UI framework
- replacing the project with a full Ant Design Pro scaffold

## Test Strategy

Use contract-style tests rather than screenshot tests:

1. shell test for dark sidebar setting + light content classes
2. visual token test for dark sidebar token and light workspace tokens
3. `/inspection/tasks` test for summary/filter/table light-surface wrappers
4. tab/header contract test through shell classes and tab container classes

## Implementation Notes

The safest repair is to make the shell explicit rather than relying on accidental default visuals. That means:

- declaring a single source of truth for light theme tokens in TypeScript for Ant Design runtime theme setup
- keeping CSS variables in `global.less` aligned with those light tokens
- narrowing dark styling to sidebar selectors only
- introducing small shared classes for summary cards, toolbars, and table shells so fixes are reusable instead of page-specific
