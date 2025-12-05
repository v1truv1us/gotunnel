# AGENTS.md - Astro.js Documentation Site

## Overview
Documentation site built with Astro 5.x for project documentation and guides.

## Development Commands
```bash
# Development
npm run dev          # Start development server
npm run build        # Build for production
npm run preview      # Preview production build

# Astro commands
npx astro dev        # Direct Astro development
npx astro build      # Direct Astro build
```

## Code Patterns
- **Astro 5.x**: Modern Astro framework with islands architecture
- **Static Generation**: Generates static HTML for documentation
- **Pages Structure**: Pages in `src/pages/` directory
- **Public Assets**: Static assets in `public/` directory

## Dependencies
- `astro`: Static site generator framework

## File Structure
- `src/pages/` - Astro page components
- `public/` - Static assets (favicon, CNAME)
- `astro.config.mjs` - Astro configuration
- `tsconfig.json` - TypeScript configuration