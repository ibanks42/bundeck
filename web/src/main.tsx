import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider, createRouter } from '@tanstack/react-router';
import ReactDOM from 'react-dom/client';
import { ThemeProvider } from './components/theme-provider';
import './main.css';
import { routeTree } from './routeTree.gen';

// Set up a Router instance
const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
});

// Register things for typesafety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}

const rootElement = document.getElementById('app');

if (!rootElement?.innerHTML) {
  const root = ReactDOM.createRoot(rootElement as HTMLElement);
  root.render(
    <Providers>
      <RouterProvider router={router} />
    </Providers>,
  );
}
const queryClient = new QueryClient();

function Providers({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider defaultTheme='system'>{children}</ThemeProvider>
    </QueryClientProvider>
  );
}
