# Optimization Summary

## Completed
- Removed generated `.js` and `.vue.js` files from `src/` to avoid source shadowing.
- Added `.gitignore` to keep build artifacts and generated JS out of the project.
- Switched Axios `baseURL` and timeout to environment-based config in `src/utils/request.ts`.
- Added reusable asset URL helper: `src/utils/file-url.ts`.
- Added reusable system-core composable: `src/composables/useSystemCore.ts`.
- Refactored `main.ts` and `layout/index.vue` to use shared system-core logic.
- Refactored `top-bar.vue` and `SystemCoreConfig.vue` to use shared asset URL helper.
- Added dynamic-route reset support in `src/router/index.ts` and used it on logout.
- Removed several debug `console.*` statements from app code.
- Verified TypeScript type check passes with `npm run type-check`.

## Not fully completed
- Full lint cleanup is still pending. The uploaded project has existing ESLint issues in several large pages.
- Full production build was not completed in this environment because the uploaded `node_modules` is missing Rollup's optional native package. Running a clean `npm install` locally should fix that.
