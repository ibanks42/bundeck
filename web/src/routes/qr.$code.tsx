import { createFileRoute } from '@tanstack/react-router';
import { toDataURL } from 'qrcode';

export const Route = createFileRoute('/qr/$code')({
  loader: ({ params: { code } }) => {
    return toDataURL(code);
  },
  component: RouteComponent,
});

function RouteComponent() {
  const data = Route.useLoaderData();
  return (
    <div className='flex  flex-col justify-center items-center mt-4 w-full'>
      <h2 className='text-3xl font-bold'>Scan QR code on your device</h2>
      <img src={data} className='size-48' alt='qr-code' />
    </div>
  );
}
