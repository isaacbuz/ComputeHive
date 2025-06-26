// Responsive breakpoints
export const breakpoints = {
  xs: 0,      // Mobile small
  sm: 390,    // Mobile
  md: 768,    // Tablet
  lg: 1024,   // Desktop
  xl: 1440,   // Desktop large
} as const;

// Container max widths
export const containerWidths = {
  sm: '100%',
  md: '720px',
  lg: '960px',
  xl: '1200px',
} as const;

// Media queries
export const mediaQueries = {
  xs: `@media (min-width: ${breakpoints.xs}px)`,
  sm: `@media (min-width: ${breakpoints.sm}px)`,
  md: `@media (min-width: ${breakpoints.md}px)`,
  lg: `@media (min-width: ${breakpoints.lg}px)`,
  xl: `@media (min-width: ${breakpoints.xl}px)`,
  
  // Max width queries
  xsMax: `@media (max-width: ${breakpoints.sm - 1}px)`,
  smMax: `@media (max-width: ${breakpoints.md - 1}px)`,
  mdMax: `@media (max-width: ${breakpoints.lg - 1}px)`,
  lgMax: `@media (max-width: ${breakpoints.xl - 1}px)`,
} as const;

// Device detection
export const devices = {
  mobile: `(max-width: ${breakpoints.md - 1}px)`,
  tablet: `(min-width: ${breakpoints.md}px) and (max-width: ${breakpoints.lg - 1}px)`,
  desktop: `(min-width: ${breakpoints.lg}px)`,
} as const;

// Touch target sizes
export const touchTargets = {
  minimum: 44, // iOS Human Interface Guidelines
  comfortable: 48,
  large: 56,
} as const;
