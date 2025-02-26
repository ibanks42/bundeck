import { FormArrayInput } from '@/components/form/form-array-input';
import { FormBooleanInput } from '@/components/form/form-boolean-input';
import { FormInput } from '@/components/form/form-input';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
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
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { useToast } from '@/hooks/use-toast';
import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery } from '@tanstack/react-query';
import { useRouter } from '@tanstack/react-router';
import parse from 'html-react-parser';
import { Loader2 } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import type { Control } from 'react-hook-form';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { Badge } from '../ui/badge';

interface Variable {
  type: string;
  default: string | number | string[] | boolean | boolean[] | number[];
  description: string;
  label: string;
}

interface PluginTemplate {
  id: string;
  title: string;
  description: string;
  file: string;
  category: string;
  label: string;
  variables: Record<string, Variable>;
}

interface CategorizedTemplates {
  [category: string]: {
    plugins: PluginTemplate[];
  };
}

interface AddPluginDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: () => void;
}

// Define a more precise variable value type
type VariableValue =
  | string
  | number
  | boolean
  | string[]
  | number[]
  | boolean[];

// Define our form values type using z.infer
type FormValues = z.infer<typeof formSchema>;

const variableSchema = z.union([
  z.string(),
  z.number(),
  z.array(z.string()),
  z.array(z.number()),
  z.boolean(),
  z.array(z.boolean()),
]);

const formSchema = z.object({
  templateId: z.string().min(1),
  variables: z.record(variableSchema),
  run_continuously: z.boolean().default(false),
  interval_seconds: z.coerce.number().min(0).default(0),
});

// Helper function to determine the input type based on the variable type
function getInputComponent(
  variable: Variable,
  control: Control<FormValues>,
  name: `variables.${string}`, // Make it a template literal type
  key: string,
) {
  // Use the variable's label if available, otherwise fall back to the key
  const label = variable.label || key.replace('BUNDECK_', '');

  // Check if it's an array type
  if (variable.type.endsWith('[]')) {
    // Extract the base type (remove the "[]" suffix)
    const baseType = variable.type.slice(0, -2);
    return (
      <FormArrayInput
        key={name}
        control={
          control as unknown as Control<
            Record<string, string | number | boolean>
          >
        }
        name={name}
        label={label}
        description={variable.description}
        type={baseType as 'text' | 'number' | 'boolean'}
      />
    );
  }

  // Handle boolean type
  if (variable.type === 'boolean') {
    return (
      <FormBooleanInput
        key={name}
        control={control as unknown as Control<Record<string, boolean>>}
        name={name}
        label={label}
        description={variable.description}
      />
    );
  }

  // Default to standard input
  return (
    <FormInput
      key={name}
      control={control}
      name={name}
      label={label}
      description={variable.description}
      type={variable.type === 'number' ? 'number' : 'text'}
    />
  );
}

export function AddPluginDialog({
  isOpen,
  onOpenChange,
  onSave,
}: AddPluginDialogProps) {
  const { toast } = useToast();
  const router = useRouter();
  const [step, setStep] = useState<'select' | 'configure'>('select');
  const [selectedTemplate, setSelectedTemplate] =
    useState<PluginTemplate | null>(null);

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      templateId: '',
      variables: {},
      run_continuously: false,
      interval_seconds: 0,
    },
  });
  const run_continuously = form.watch('run_continuously');

  // Query for fetching templates
  const { data: templates, isError } = useQuery({
    queryKey: ['plugin-templates'],
    queryFn: async () => {
      const response = await fetch('/api/plugins/templates');
      if (!response.ok) {
        throw new Error('Failed to load plugin templates');
      }
      return response.json() as Promise<PluginTemplate[]>;
    },
    enabled: isOpen,
  });

  // Transform templates into categories and ensure they're in the order from list.json
  const categorizedTemplates = useMemo(() => {
    if (!templates) {
      return {
        categories: [] as string[],
        categoryMap: {} as Record<string, { plugins: PluginTemplate[] }>,
      };
    }

    // Create a structure to track category order and plugins
    const result: {
      categories: string[];
      categoryMap: Record<string, { plugins: PluginTemplate[] }>;
    } = {
      categories: [],
      categoryMap: {},
    };

    // List of known categories in the desired order
    // (matching the order in list.json)
    const knownCategoryOrder = ['OBS', 'Operating System', 'Input'];

    // First, group templates by category
    for (const template of templates) {
      const category = template.category;

      if (!result.categoryMap[category]) {
        // Add to categories list if first time seeing this category
        if (!result.categories.includes(category)) {
          result.categories.push(category);
        }
        result.categoryMap[category] = { plugins: [] };
      }

      // Add the template to its category
      result.categoryMap[category].plugins.push(template);
    }

    // Sort the categories according to the known order
    result.categories.sort((a, b) => {
      const indexA = knownCategoryOrder.indexOf(a);
      const indexB = knownCategoryOrder.indexOf(b);

      // If both categories are in our known list, sort by their order
      if (indexA !== -1 && indexB !== -1) {
        return indexA - indexB;
      }

      // If only one is in the list, the one in the list comes first
      if (indexA !== -1) return -1;
      if (indexB !== -1) return 1;

      // If neither is in the list, keep original order
      return 0;
    });

    return result;
  }, [templates]);

  // Mutation for creating plugin
  const { mutate, isPending } = useMutation({
    mutationFn: async (values: FormValues) => {
      const response = await fetch('/api/plugins/templates/create', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...values,
          run_continuously: values.run_continuously,
          interval_seconds: values.interval_seconds,
        }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Failed to create plugin');
      }

      return response.json();
    },
    onSuccess: () => {
      toast({
        title: 'Success',
        description: 'Plugin added successfully',
      });
      onSave();
      onOpenChange(false);
      setStep('select');
      setSelectedTemplate(null);
      form.reset();
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

  const handleTemplateSelect = (template: PluginTemplate) => {
    setSelectedTemplate(template);
    // Initialize form with template values
    const initialVars: Record<string, VariableValue> = {};
    for (const [key, variable] of Object.entries(template.variables)) {
      initialVars[key] = variable.default;
    }
    form.reset({
      templateId: template.id,
      variables: initialVars,
    });
    setStep('configure');
  };

  function onSubmit(values: FormValues) {
    mutate(values);
  }

  // Reset form when dialog closes
  useEffect(() => {
    if (!isOpen) {
      setStep('select');
      setSelectedTemplate(null);
      form.reset();
    }
  }, [isOpen, form]);

  if (isError) {
    return (
      <Dialog open={isOpen} onOpenChange={onOpenChange}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Error</DialogTitle>
            <DialogDescription>
              Failed to load plugin templates
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button onClick={() => onOpenChange(false)}>Close</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-[550px] max-h-[90vh] flex flex-col'>
        <DialogHeader className='flex-shrink-0'>
          <DialogTitle>Add Plugin</DialogTitle>
          <DialogDescription>
            Create a new plugin from a template or write your own.
          </DialogDescription>
        </DialogHeader>

        {step === 'select' && (
          <div className='flex flex-col gap-4 overflow-hidden flex-1'>
            {isError && (
              <div className='text-destructive'>
                Failed to load plugin templates
              </div>
            )}
            {!templates ? (
              <div className='flex h-[300px] items-center justify-center'>
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                Loading templates...
              </div>
            ) : (
              <div className='overflow-auto rounded-md border bg-background flex-1'>
                <Accordion type='multiple' className='w-full'>
                  {categorizedTemplates.categories.map((category) => (
                    <AccordionItem value={category} key={category}>
                      <AccordionTrigger className='px-4'>
                        {category}
                      </AccordionTrigger>
                      <AccordionContent>
                        <div className='flex flex-col gap-2 p-2'>
                          {categorizedTemplates.categoryMap[
                            category
                          ].plugins.map((template) => (
                            <Button
                              key={template.id}
                              variant='outline'
                              className='w-full h-auto py-3 justify-start text-left'
                              onClick={() => handleTemplateSelect(template)}
                            >
                              <div className='relative flex items-center gap-3 w-full'>
                                <div className='flex flex-col'>
                                  <div className='font-medium'>
                                    {template.title}
                                  </div>
                                  <div className='text-sm text-muted-foreground'>
                                    {parse(template.description)}
                                  </div>
                                </div>
                                {template.label && (
                                  <Badge
                                    variant='outline'
                                    className='absolute top-2 right-2'
                                  >
                                    {template.label}
                                  </Badge>
                                )}
                              </div>
                            </Button>
                          ))}
                        </div>
                      </AccordionContent>
                    </AccordionItem>
                  ))}
                </Accordion>
              </div>
            )}
          </div>
        )}

        {step === 'configure' && selectedTemplate && (
          <Form {...form}>
            <form
              onSubmit={form.handleSubmit(onSubmit)}
              className='flex flex-col flex-1 overflow-hidden'
            >
              <div className='overflow-y-auto py-4 flex-1'>
                <div className='grid gap-4 px-2'>
                  {Object.entries(selectedTemplate.variables).map(
                    ([key, variable]) =>
                      getInputComponent(
                        variable,
                        form.control,
                        `variables.${key}`,
                        key.replace('BUNDECK_', ''),
                      ),
                  )}

                  {/* Add continuous run settings */}
                  <div className='border rounded-md p-4'>
                    <h3 className='font-medium mb-2'>Run Settings</h3>
                    <div className='space-y-4'>
                      <FormField
                        control={form.control}
                        name='run_continuously'
                        render={({ field }) => (
                          <FormItem className='flex flex-row items-start space-x-3 space-y-0'>
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

                      {run_continuously && (
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
                      )}
                    </div>
                  </div>
                </div>
              </div>
              <DialogFooter className='flex-shrink-0 pt-2'>
                <Button
                  type='button'
                  variant='outline'
                  onClick={() => {
                    setStep('select');
                    setSelectedTemplate(null);
                    form.reset();
                  }}
                >
                  Back
                </Button>
                <Button type='submit' disabled={isPending}>
                  {isPending && (
                    <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  )}
                  Add Plugin
                </Button>
              </DialogFooter>
            </form>
          </Form>
        )}
      </DialogContent>
    </Dialog>
  );
}
