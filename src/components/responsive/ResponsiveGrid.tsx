import React from 'react';
import { Grid, GridProps } from '@mui/material';
import { useResponsive, useResponsiveStyles } from '../../hooks/useResponsive';

interface ResponsiveGridProps extends Omit<GridProps, 'item' | 'container'> {
  children: React.ReactNode;
  className?: string;
  columns?: {
    mobile?: number;
    tablet?: number;
    desktop?: number;
    wide?: number;
  };
  gap?: {
    mobile?: string;
    tablet?: string;
    desktop?: string;
    wide?: string;
  };
  alignItems?: 'start' | 'center' | 'end' | 'stretch';
  justifyItems?: 'start' | 'center' | 'end' | 'stretch';
  as?: keyof JSX.IntrinsicElements;
}

interface ResponsiveGridItemProps extends Omit<GridProps, 'item' | 'container'> {
  children: React.ReactNode;
  size?: {
    xs?: number;
    sm?: number;
    md?: number;
    lg?: number;
    xl?: number;
  };
}

export const ResponsiveGrid: React.FC<ResponsiveGridProps> = ({
  children,
  className = '',
  columns,
  gap,
  alignItems = 'stretch',
  justifyItems = 'start',
  as: Component = 'div',
}) => {
  const { grid } = useResponsiveStyles();

  const getColumns = () => {
    if (columns) {
      // Use custom column configuration
      return {
        mobile: columns.mobile || 1,
        tablet: columns.tablet || 2,
        desktop: columns.desktop || 3,
        wide: columns.wide || 4,
      };
    }
    
    // Use default responsive columns
    return {
      mobile: 1,
      tablet: 2,
      desktop: 3,
      wide: 4,
    };
  };

  const getGap = () => {
    if (gap) {
      return {
        mobile: gap.mobile || '1rem',
        tablet: gap.tablet || '1.5rem',
        desktop: gap.desktop || '2rem',
        wide: gap.wide || '2rem',
      };
    }
    
    return {
      mobile: '1rem',
      tablet: '1.5rem',
      desktop: '2rem',
      wide: '2rem',
    };
  };

  const columnConfig = getColumns();
  const gapConfig = getGap();

  const gridStyles: React.CSSProperties = {
    display: 'grid',
    gridTemplateColumns: `repeat(${columnConfig.mobile}, 1fr)`,
    gap: gapConfig.mobile,
    alignItems,
    justifyItems,
    width: '100%',
  };

  // CSS custom properties for responsive behavior
  const cssVars = {
    '--grid-columns-mobile': columnConfig.mobile,
    '--grid-columns-tablet': columnConfig.tablet,
    '--grid-columns-desktop': columnConfig.desktop,
    '--grid-columns-wide': columnConfig.wide,
    '--grid-gap-mobile': gapConfig.mobile,
    '--grid-gap-tablet': gapConfig.tablet,
    '--grid-gap-desktop': gapConfig.desktop,
    '--grid-gap-wide': gapConfig.wide,
  } as React.CSSProperties;

  return (
    <Component
      className={`responsive-grid ${className}`}
      style={{ ...gridStyles, ...cssVars }}
    >
      {children}
    </Component>
  );
};

export const ResponsiveGridItem: React.FC<ResponsiveGridItemProps> = ({
  children,
  size = { xs: 12, sm: 6, md: 4, lg: 3, xl: 3 },
  ...props
}) => {
  return (
    <Grid
      item
      xs={size.xs}
      sm={size.sm}
      md={size.md}
      lg={size.lg}
      xl={size.xl}
      {...props}
    >
      {children}
    </Grid>
  );
};

// Preset grid layouts
export const GridLayouts = {
  // Single column on mobile, 2 on tablet, 3 on desktop
  cards: {
    xs: 12,
    sm: 6,
    md: 4,
    lg: 4,
    xl: 3,
  },
  
  // Full width on mobile, half on tablet and up
  twoColumn: {
    xs: 12,
    sm: 6,
    md: 6,
    lg: 6,
    xl: 6,
  },
  
  // Full width on mobile/tablet, third on desktop
  threeColumn: {
    xs: 12,
    sm: 12,
    md: 4,
    lg: 4,
    xl: 4,
  },
  
  // Sidebar layout: 1/3 sidebar, 2/3 content
  sidebar: {
    xs: 12,
    sm: 12,
    md: 4,
    lg: 3,
    xl: 3,
  },
  
  content: {
    xs: 12,
    sm: 12,
    md: 8,
    lg: 9,
    xl: 9,
  },
  
  // Dashboard widgets
  widget: {
    xs: 12,
    sm: 6,
    md: 6,
    lg: 3,
    xl: 3,
  },
  
  // Form layouts
  formFull: {
    xs: 12,
    sm: 12,
    md: 12,
    lg: 12,
    xl: 12,
  },
  
  formHalf: {
    xs: 12,
    sm: 12,
    md: 6,
    lg: 6,
    xl: 6,
  },
};

// Responsive Masonry Grid for variable height items
interface ResponsiveMasonryProps {
  children: React.ReactNode;
  columns?: {
    xs?: number;
    sm?: number;
    md?: number;
    lg?: number;
    xl?: number;
  };
  spacing?: number;
}

export const ResponsiveMasonry: React.FC<ResponsiveMasonryProps> = ({
  children,
  columns = { xs: 1, sm: 2, md: 3, lg: 4, xl: 4 },
  spacing = 2,
}) => {
  const { breakpoint } = useResponsive();
  const columnCount = columns[breakpoint] || 1;

  return (
    <div
      style={{
        columnCount,
        columnGap: spacing * 8,
      }}
    >
      {React.Children.map(children, (child, index) => (
        <div
          key={index}
          style={{
            breakInside: 'avoid',
            marginBottom: spacing * 8,
          }}
        >
          {child}
        </div>
      ))}
    </div>
  );
};

// Specialized grid variants
export const AutoGrid: React.FC<Omit<ResponsiveGridProps, 'columns'>> = (props) => (
  <ResponsiveGrid
    {...props}
    columns={{
      mobile: 1,
      tablet: 2,
      desktop: 3,
      wide: 4,
    }}
    className={`auto-grid ${props.className || ''}`}
  />
);

export const CardGrid: React.FC<Omit<ResponsiveGridProps, 'columns' | 'gap'>> = (props) => (
  <ResponsiveGrid
    {...props}
    columns={{
      mobile: 1,
      tablet: 2,
      desktop: 3,
      wide: 4,
    }}
    gap={{
      mobile: '1rem',
      tablet: '1.5rem',
      desktop: '2rem',
      wide: '2rem',
    }}
    className={`card-grid ${props.className || ''}`}
  />
);

export const FormGrid: React.FC<Omit<ResponsiveGridProps, 'columns' | 'gap'>> = (props) => (
  <ResponsiveGrid
    {...props}
    columns={{
      mobile: 1,
      tablet: 2,
      desktop: 2,
      wide: 3,
    }}
    gap={{
      mobile: '1rem',
      tablet: '1.5rem',
      desktop: '2rem',
      wide: '2rem',
    }}
    className={`form-grid ${props.className || ''}`}
  />
);

export const DashboardGrid: React.FC<Omit<ResponsiveGridProps, 'columns' | 'gap'>> = (props) => (
  <ResponsiveGrid
    {...props}
    columns={{
      mobile: 1,
      tablet: 2,
      desktop: 4,
      wide: 6,
    }}
    gap={{
      mobile: '1rem',
      tablet: '1.5rem',
      desktop: '1.5rem',
      wide: '2rem',
    }}
    className={`dashboard-grid ${props.className || ''}`}
  />
);

export default ResponsiveGrid; 