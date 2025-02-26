import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useToast } from '@/hooks/use-toast';
import type { Plugin } from '@/types/plugin';
import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery } from '@tanstack/react-query';
import { useRouter } from '@tanstack/react-router';
import { ImageIcon, Loader2 } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { FormEditor } from '../form/form-editor';
import { FormInput } from '../form/form-input';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '../ui/form';
import { Input } from '../ui/input';

interface EditPluginDialogProps {
  plugin?: Plugin;
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: () => void;
}

const schema = z.object({
  name: z.string().min(1),
  code: z.string().min(1),
  image: z.instanceof(File).optional(),
  run_continuously: z.boolean().default(false),
  interval_seconds: z.coerce.number().min(0).default(0),
});

export function EditPluginDialog({
  plugin,
  isOpen,
  onOpenChange,
  onSave,
}: EditPluginDialogProps) {
  const defaultValues = {
    name: plugin?.name ?? '',
    code:
      plugin?.code ??
      `// Example plugin using Bun
import { v4 as uuidv4 } from 'uuid';

// You can show output to the user using console.log, console.error, etc.
console.log(uuidv4());`,
    image: undefined,
    run_continuously: plugin?.run_continuously ?? false,
    interval_seconds: plugin?.interval_seconds ?? 0,
  };
  const router = useRouter();
  const { toast } = useToast();
  const [selectedImage, setSelectedImage] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues,
  });
  const run_continuously = form.watch('run_continuously');

  const { data: image, isSuccess: imageLoaded } = useQuery({
    queryKey: ['plugin-image', plugin?.id],
    queryFn: async () => {
      const response = await fetch(`/api/plugins/${plugin?.id}/image`);
      if (!response.ok) {
        return null;
      }
      return response.blob();
    },
    retry: false,
    enabled: !!plugin?.id,
  });

  // Handle initial image load from query
  useEffect(() => {
    if (imageLoaded && image) {
      const imageFile = new File([image], 'plugin-image', { type: image.type });
      setSelectedImage(imageFile);
      setPreviewUrl(URL.createObjectURL(image));
      form.setValue('image', imageFile, { shouldValidate: true });
    }
  }, [imageLoaded, image, form]);

  // Reset form when dialog opens.
  useEffect(() => {
    if (!isOpen) return;

    if (plugin) {
      form.reset({
        name: plugin.name,
        code: plugin.code,
        run_continuously: plugin.run_continuously,
        interval_seconds: plugin.interval_seconds,
      });
      // Do not clear image state when editing an existing plugin.
    } else {
      form.reset({
        name: '',
        code: `// Example plugin using Bun
import { v4 as uuidv4 } from 'uuid';

// You can show output to the user using console.log, console.error, etc.
console.log(uuidv4());`,
        image: undefined,
        run_continuously: false,
        interval_seconds: 0,
      });
      setPreviewUrl(null);
      setSelectedImage(null);
    }

    // Reset file input
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  }, [plugin, isOpen, form]);

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      if (!file.type.startsWith('image/')) {
        toast({
          title: 'Error',
          description: 'Please select an image file',
          variant: 'destructive',
        });
        if (fileInputRef.current) {
          fileInputRef.current.value = '';
        }
        return;
      }
      setSelectedImage(file);
      setPreviewUrl(URL.createObjectURL(file));
      form.setValue('image', file, { shouldValidate: true });
    }
    // Do nothing on cancel. Do not clear the image state.
  };

  const handleClearImage = () => {
    setSelectedImage(null);
    setPreviewUrl(null);
    form.setValue('image', undefined, { shouldValidate: true });
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  // Cleanup preview URLs on unmount
  useEffect(() => {
    return () => {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
      }
    };
  }, [previewUrl]);

  const { mutate, isPending: isUpdating } = useMutation({
    mutationFn: async (values: z.infer<typeof schema>) => {
      const formData = new FormData();
      formData.append('name', values.name);
      formData.append('code', values.code);
      formData.append('run_continuously', values.run_continuously.toString());
      formData.append('interval_seconds', values.interval_seconds.toString());

      if (selectedImage) {
        formData.append('image', selectedImage);
      }

      if (plugin) {
        // Update existing plugin
        formData.append('order_num', plugin.order_num.toString());
        const response = await fetch(`/api/plugins/${plugin.id}/code`, {
          method: 'PUT',
          body: formData,
        });
        const data = await response.json();
        if (data.error) {
          throw new Error(data.error);
        }
        return data;
      }

      // Create new plugin
      formData.append('order_num', '999'); // Will be last in order
      const response = await fetch('/api/plugins', {
        method: 'POST',
        body: formData,
      });
      const data = await response.json();
      if (data.error) {
        throw new Error(data.error);
      }
      return data;
    },
    onSuccess: () => {
      onSave();
      onOpenChange(false);
      toast({
        title: 'Success',
        description: 'Plugin saved successfully',
      });
      router.invalidate();
    },
    onError: (error) => {
      toast({
        title: 'Error',
        description: error.message,
        variant: 'destructive',
      });
    },
  });

  function onSubmit(values: z.infer<typeof schema>) {
    mutate(values);
  }

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent
        className='max-w-3xl'
        onEscapeKeyDown={(e) => e.preventDefault()}
      >
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <DialogHeader>
              <DialogTitle>
                {plugin ? 'Edit Plugin' : 'Add New Plugin'}
              </DialogTitle>
              <DialogDescription>
                {plugin
                  ? 'Modify your plugin code, name, and image below.'
                  : 'Create a new plugin using TypeScript/JavaScript. You can use any npm package!'}
              </DialogDescription>
            </DialogHeader>
            <div className='grid gap-4 py-4'>
              <div className='grid gap-2'>
                <FormInput
                  control={form.control}
                  name='name'
                  label='Name'
                  placeholder='Plugin name'
                />
              </div>
              <div className='grid gap-2'>
                <div className='flex items-center gap-4'>
                  <div className='w-32 h-32 rounded-lg border flex items-center justify-center overflow-hidden'>
                    {previewUrl ? (
                      <img
                        src={previewUrl}
                        alt='Plugin icon preview'
                        className='w-full h-full object-contain'
                      />
                    ) : (
                      <ImageIcon className='w-12 h-12 text-muted-foreground' />
                    )}
                  </div>
                  <div className='flex-1'>
                    <FormLabel>Image</FormLabel>
                    <Input
                      type='file'
                      accept='image/*'
                      onChange={handleImageChange}
                      className='w-full'
                      ref={fileInputRef}
                    />
                    {previewUrl && (
                      <Button
                        type='button'
                        variant='destructive'
                        className='mt-2'
                        onClick={handleClearImage}
                        tabIndex={0}
                        aria-label='Clear image'
                      >
                        Clear Image
                      </Button>
                    )}
                    <p className='text-sm text-muted-foreground mt-1'>
                      Upload an image for your plugin (optional)
                    </p>
                  </div>
                </div>
              </div>

              {/* Add continuous run settings */}
              <div className='grid grid-cols-2 gap-4'>
                <FormField
                  control={form.control}
                  name='run_continuously'
                  render={({ field }) => (
                    <FormItem className='flex flex-row items-start space-x-3 space-y-0 rounded-md border p-4'>
                      <FormControl>
                        <Checkbox
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      </FormControl>
                      <div className='space-y-1 leading-none'>
                        <FormLabel>Run Continuously</FormLabel>
                        <FormDescription>
                          Run this plugin at regular intervals
                        </FormDescription>
                      </div>
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name='interval_seconds'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Interval (seconds)</FormLabel>
                      <FormControl>
                        <Input
                          type='number'
                          min={0}
                          {...field}
                          onChange={field.onChange}
                        />
                      </FormControl>
                      <FormDescription>
                        How often to run the plugin (in seconds)
                      </FormDescription>
                    </FormItem>
                  )}
                />
              </div>

              <div className='h-[400px]'>
                <FormEditor
                  control={form.control}
                  name='code'
                  description='Your plugin code'
                  containerClassName='h-[400px]'
                />
              </div>
            </div>
            <DialogFooter>
              <Button
                type='button'
                variant='outline'
                onClick={() => onOpenChange(false)}
              >
                Cancel
              </Button>
              <Button type='submit'>
                {isUpdating && (
                  <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                )}
                {plugin ? 'Update Plugin' : 'Create Plugin'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
