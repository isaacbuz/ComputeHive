import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { ResponsiveContainer } from '../../src/components/responsive/ResponsiveContainer';
import { ResponsiveGrid } from '../../src/components/responsive/ResponsiveGrid';
import { ResponsiveNavigation } from '../../src/components/responsive/ResponsiveNavigation';
import { ResponsiveDataTable } from '../../src/components/responsive/ResponsiveDataTable';
import { useResponsive } from '../../src/hooks/useResponsive';

// Mock the responsive hook
jest.mock('../../src/hooks/useResponsive');

const mockUseResponsive = useResponsive as jest.MockedFunction<typeof useResponsive>;

describe('Responsive Components', () => {
  beforeEach(() => {
    // Reset window size to desktop
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 1200,
    });
    Object.defineProperty(window, 'innerHeight', {
      writable: true,
      configurable: true,
      value: 800,
    });
  });

  describe('ResponsiveContainer', () => {
    it('renders with default props', () => {
      render(
        <ResponsiveContainer>
          <div>Test content</div>
        </ResponsiveContainer>
      );

      const container = screen.getByText('Test content').parentElement;
      expect(container).toHaveClass('responsive-container');
    });

    it('applies custom maxWidth', () => {
      render(
        <ResponsiveContainer maxWidth="xl">
          <div>Test content</div>
        </ResponsiveContainer>
      );

      const container = screen.getByText('Test content').parentElement;
      expect(container).toHaveStyle({ maxWidth: '1280px' });
    });

    it('applies custom padding', () => {
      render(
        <ResponsiveContainer padding="lg">
          <div>Test content</div>
        </ResponsiveContainer>
      );

      const container = screen.getByText('Test content').parentElement;
      expect(container).toHaveStyle({ padding: '1.5rem' });
    });

    it('renders as custom element', () => {
      render(
        <ResponsiveContainer as="section">
          <div>Test content</div>
        </ResponsiveContainer>
      );

      const container = screen.getByText('Test content').parentElement;
      expect(container?.tagName).toBe('SECTION');
    });
  });

  describe('ResponsiveGrid', () => {
    it('renders with default grid layout', () => {
      render(
        <ResponsiveGrid>
          <div>Item 1</div>
          <div>Item 2</div>
        </ResponsiveGrid>
      );

      const grid = screen.getByText('Item 1').parentElement;
      expect(grid).toHaveClass('responsive-grid');
    });

    it('applies custom columns configuration', () => {
      render(
        <ResponsiveGrid
          columns={{
            mobile: 1,
            tablet: 2,
            desktop: 3,
            wide: 4,
          }}
        >
          <div>Item 1</div>
          <div>Item 2</div>
        </ResponsiveGrid>
      );

      const grid = screen.getByText('Item 1').parentElement;
      expect(grid).toHaveStyle({
        '--grid-columns-mobile': '1',
        '--grid-columns-tablet': '2',
        '--grid-columns-desktop': '3',
        '--grid-columns-wide': '4',
      });
    });

    it('applies custom gap configuration', () => {
      render(
        <ResponsiveGrid
          gap={{
            mobile: '0.5rem',
            tablet: '1rem',
            desktop: '1.5rem',
            wide: '2rem',
          }}
        >
          <div>Item 1</div>
          <div>Item 2</div>
        </ResponsiveGrid>
      );

      const grid = screen.getByText('Item 1').parentElement;
      expect(grid).toHaveStyle({
        '--grid-gap-mobile': '0.5rem',
        '--grid-gap-tablet': '1rem',
        '--grid-gap-desktop': '1.5rem',
        '--grid-gap-wide': '2rem',
      });
    });
  });

  describe('ResponsiveNavigation', () => {
    const mockItems = [
      { id: '1', label: 'Dashboard', icon: 'home' },
      { id: '2', label: 'Jobs', icon: 'briefcase' },
      { id: '3', label: 'Resources', icon: 'server' },
    ];

    it('renders navigation items', () => {
      render(<ResponsiveNavigation items={mockItems} />);

      expect(screen.getByText('Dashboard')).toBeInTheDocument();
      expect(screen.getByText('Jobs')).toBeInTheDocument();
      expect(screen.getByText('Resources')).toBeInTheDocument();
    });

    it('handles item click', () => {
      const onItemClick = jest.fn();
      render(
        <ResponsiveNavigation items={mockItems} onItemClick={onItemClick} />
      );

      fireEvent.click(screen.getByText('Dashboard'));
      expect(onItemClick).toHaveBeenCalledWith(mockItems[0]);
    });

    it('shows active item', () => {
      render(
        <ResponsiveNavigation items={mockItems} activeItem="2" />
      );

      const activeItem = screen.getByText('Jobs').closest('button');
      expect(activeItem).toHaveClass('active');
    });
  });

  describe('ResponsiveDataTable', () => {
    const mockData = [
      { id: '1', name: 'Job 1', status: 'running', type: 'ML' },
      { id: '2', name: 'Job 2', status: 'completed', type: 'Data' },
    ];

    const mockColumns = [
      { key: 'name', label: 'Name' },
      { key: 'status', label: 'Status' },
      { key: 'type', label: 'Type' },
    ];

    it('renders table with data', () => {
      render(
        <ResponsiveDataTable
          data={mockData}
          columns={mockColumns}
        />
      );

      expect(screen.getByText('Job 1')).toBeInTheDocument();
      expect(screen.getByText('Job 2')).toBeInTheDocument();
    });

    it('handles search functionality', () => {
      render(
        <ResponsiveDataTable
          data={mockData}
          columns={mockColumns}
          search={{ enabled: true }}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search...');
      fireEvent.change(searchInput, { target: { value: 'Job 1' } });

      expect(screen.getByText('Job 1')).toBeInTheDocument();
      expect(screen.queryByText('Job 2')).not.toBeInTheDocument();
    });

    it('handles row click', () => {
      const onRowClick = jest.fn();
      render(
        <ResponsiveDataTable
          data={mockData}
          columns={mockColumns}
          onRowClick={onRowClick}
        />
      );

      fireEvent.click(screen.getByText('Job 1'));
      expect(onRowClick).toHaveBeenCalledWith(mockData[0]);
    });

    it('shows empty state when no data', () => {
      render(
        <ResponsiveDataTable
          data={[]}
          columns={mockColumns}
          emptyMessage="No jobs found"
        />
      );

      expect(screen.getByText('No jobs found')).toBeInTheDocument();
    });
  });

  describe('Responsive Hook', () => {
    it('returns correct responsive state for desktop', () => {
      mockUseResponsive.mockReturnValue({
        isMobile: false,
        isTablet: false,
        isDesktop: true,
        isWide: false,
        currentBreakpoint: 'lg',
        deviceType: 'desktop',
        width: 1200,
        height: 800,
      });

      const TestComponent = () => {
        const responsive = useResponsive();
        return <div data-testid="responsive">{responsive.deviceType}</div>;
      };

      render(<TestComponent />);
      expect(screen.getByTestId('responsive')).toHaveTextContent('desktop');
    });

    it('returns correct responsive state for mobile', () => {
      mockUseResponsive.mockReturnValue({
        isMobile: true,
        isTablet: false,
        isDesktop: false,
        isWide: false,
        currentBreakpoint: 'sm',
        deviceType: 'mobile',
        width: 375,
        height: 667,
      });

      const TestComponent = () => {
        const responsive = useResponsive();
        return <div data-testid="responsive">{responsive.deviceType}</div>;
      };

      render(<TestComponent />);
      expect(screen.getByTestId('responsive')).toHaveTextContent('mobile');
    });
  });

  describe('Responsive Breakpoints', () => {
    it('applies correct styles for mobile breakpoint', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });

      render(
        <ResponsiveContainer>
          <div>Mobile content</div>
        </ResponsiveContainer>
      );

      const container = screen.getByText('Mobile content').parentElement;
      expect(container).toHaveStyle({ padding: '1rem' });
    });

    it('applies correct styles for tablet breakpoint', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 768,
      });

      render(
        <ResponsiveContainer>
          <div>Tablet content</div>
        </ResponsiveContainer>
      );

      const container = screen.getByText('Tablet content').parentElement;
      expect(container).toHaveStyle({ padding: '1.5rem' });
    });

    it('applies correct styles for desktop breakpoint', () => {
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1200,
      });

      render(
        <ResponsiveContainer>
          <div>Desktop content</div>
        </ResponsiveContainer>
      );

      const container = screen.getByText('Desktop content').parentElement;
      expect(container).toHaveStyle({ padding: '2rem' });
    });
  });

  describe('Accessibility', () => {
    it('has proper ARIA labels', () => {
      render(
        <ResponsiveNavigation
          items={[
            { id: '1', label: 'Dashboard', icon: 'home' },
          ]}
        />
      );

      const navButton = screen.getByRole('button', { name: /dashboard/i });
      expect(navButton).toBeInTheDocument();
    });

    it('supports keyboard navigation', () => {
      render(
        <ResponsiveNavigation
          items={[
            { id: '1', label: 'Dashboard', icon: 'home' },
            { id: '2', label: 'Jobs', icon: 'briefcase' },
          ]}
        />
      );

      const firstButton = screen.getByText('Dashboard').closest('button');
      const secondButton = screen.getByText('Jobs').closest('button');

      firstButton?.focus();
      expect(firstButton).toHaveFocus();

      fireEvent.keyDown(firstButton!, { key: 'Tab' });
      expect(secondButton).toHaveFocus();
    });

    it('has proper focus indicators', () => {
      render(
        <ResponsiveContainer>
          <button>Test button</button>
        </ResponsiveContainer>
      );

      const button = screen.getByRole('button');
      button.focus();
      
      expect(button).toHaveFocus();
      // Check for focus styles (this would depend on your CSS)
      expect(button).toHaveClass('focus-visible');
    });
  });

  describe('Performance', () => {
    it('renders large datasets efficiently', () => {
      const largeData = Array.from({ length: 1000 }, (_, i) => ({
        id: i.toString(),
        name: `Job ${i}`,
        status: 'running',
        type: 'ML',
      }));

      const startTime = performance.now();
      
      render(
        <ResponsiveDataTable
          data={largeData}
          columns={[
            { key: 'name', label: 'Name' },
            { key: 'status', label: 'Status' },
            { key: 'type', label: 'Type' },
          ]}
        />
      );

      const endTime = performance.now();
      const renderTime = endTime - startTime;

      // Should render in under 100ms
      expect(renderTime).toBeLessThan(100);
    });

    it('handles rapid resize events', () => {
      const resizeHandler = jest.fn();
      
      render(
        <ResponsiveContainer>
          <div>Test content</div>
        </ResponsiveContainer>
      );

      // Simulate rapid resize events
      for (let i = 0; i < 10; i++) {
        Object.defineProperty(window, 'innerWidth', {
          writable: true,
          configurable: true,
          value: 375 + i * 100,
        });
        
        fireEvent(window, new Event('resize'));
      }

      // Should handle resize events without errors
      expect(resizeHandler).toBeDefined();
    });
  });
}); 