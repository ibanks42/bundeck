import { AddPluginDialog } from '@/components/plugins/add-plugin-dialog';
import { EditPluginDialog } from '@/components/plugins/edit-dialog';
import { SortablePluginCard } from '@/components/plugins/plugin-card';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useConfirmDialog } from '@/hooks/use-confirm-dialog';
import { useToast } from '@/hooks/use-toast';

import type { Plugin } from '@/types/plugin';
import {
  DndContext,
  type DragEndEvent,
  MouseSensor,
  TouchSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import {
  SortableContext,
  arrayMove,
  rectSortingStrategy,
} from '@dnd-kit/sortable';
import { useMutation } from '@tanstack/react-query';
import { createFileRoute, useRouter } from '@tanstack/react-router';
import {
  Blocks,
  FullscreenIcon,
  PencilIcon,
  Plug2Icon,
  PlugZap2Icon,
  RefreshCwIcon,
} from 'lucide-react';
import { useEffect, useState } from 'react';

export const Route = createFileRoute('/')({
  component: HomeComponent,
  loader: async () => {
    const response = await fetch('/api/plugins');
    const plugins = (await response.json()) as Plugin[];
    return plugins;
  },
});

function HomeComponent() {
  const plugins = Route.useLoaderData();
  const router = useRouter();
  const { toast } = useToast();
  const [pluginsCopy, setPluginsCopy] = useState(plugins);
  const [isEditMode, setIsEditMode] = useState(false);

  const [editingPlugin, setEditingPlugin] = useState<Plugin | undefined>(
    undefined,
  );
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);

  const { confirm: confirmDelete, ConfirmDialog: DeleteConfirmDialog } =
    useConfirmDialog({
      title: 'Delete Plugin',
      description:
        'Are you sure you want to delete this plugin? This action cannot be undone.',
      confirmText: 'Delete',
      confirmVariant: 'destructive',
    });

  const mouseSensor = useSensor(MouseSensor, {
    activationConstraint: {
      distance: 40,
    },
  });
  const touchSensor = useSensor(TouchSensor, {
    activationConstraint: {
      distance: 40,
    },
  });
  const sensors = useSensors(mouseSensor, touchSensor);

  const { mutate: setPlugins } = useMutation({
    mutationFn: (plugins: Pick<Plugin, 'id' | 'order_num'>[]) => {
      return fetch('/api/plugins/reorder', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(plugins),
      });
    },
    onSuccess: () => {
      router.invalidate();
    },
    onError: (error) => {
      console.error('Error updating plugin order:', error);
      toast({
        title: 'Error',
        description: 'Error updating plugin order',
        variant: 'destructive',
      });
    },
  });

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event;

    if (over && active.id !== over.id) {
      const oldIndex = pluginsCopy.findIndex((item) => item.id === active.id);
      const newIndex = pluginsCopy.findIndex((item) => item.id === over.id);

      const updated = arrayMove(pluginsCopy, oldIndex, newIndex);

      const updatedOrders = updated.map((plugin, index) => ({
        id: plugin.id,
        order_num: index,
      }));

      setPluginsCopy(updated);
      setPlugins(updatedOrders);
      return updated;
    }
  };

  const handleEdit = (plugin: Plugin) => {
    setEditingPlugin(plugin);
    setIsEditDialogOpen(true);
  };

  const handleDelete = (plugin: Plugin) => {
    confirmDelete(async () => {
      try {
        await fetch(`/api/plugins/${plugin.id}`, {
          method: 'DELETE',
        });
        router.invalidate();
      } catch (error) {
        console.error('Error deleting plugin:', error);
        toast({
          title: 'Error',
          description: 'Error deleting plugin',
          variant: 'destructive',
        });
      }
    });
  };

  const handleSave = async () => {
    router.invalidate();
  };

  useEffect(() => {
    setPluginsCopy(plugins);
  }, [plugins]);

  return (
    <div className='container mx-auto p-6'>
      <div className='flex justify-between items-center mb-6'>
        <h1 className='text-3xl font-bold'>BunDeck</h1>
        <div className='flex flex-col gap-2 md:flex-row'>
          <Button
            variant='outline'
            onClick={() => {
              window.document.location = '.';
            }}
          >
            <RefreshCwIcon className='size-4 mr-2' />
            Refresh
          </Button>
          <Button
            variant='outline'
            onClick={() => {
              // toggle fullscreen
              if (document.fullscreenElement) {
                document.exitFullscreen();
              } else {
                document.documentElement.requestFullscreen();
              }
            }}
          >
            <FullscreenIcon className='size-4 mr-2' />
            Fullscreen
          </Button>
          <Button variant='outline' onClick={() => setIsEditMode(!isEditMode)}>
            <PencilIcon className='size-4 mr-2' />
            {isEditMode ? 'Done' : 'Edit'}
          </Button>
          {isEditMode && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant='outline'>
                  <Plug2Icon />
                  Add Plugins
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuItem onClick={() => setIsAddDialogOpen(true)}>
                  <PlugZap2Icon />
                  Pre-made Plugin
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => {
                    setEditingPlugin(undefined);
                    setIsEditDialogOpen(true);
                  }}
                >
                  <Blocks />
                  Custom Plugin
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </div>

      {pluginsCopy &&
        (isEditMode ? (
          <DndContext sensors={sensors} onDragEnd={handleDragEnd}>
            <SortableContext items={pluginsCopy} strategy={rectSortingStrategy}>
              <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'>
                {pluginsCopy.map((plugin) => (
                  <SortablePluginCard
                    key={plugin.id}
                    plugin={plugin}
                    onEdit={handleEdit}
                    onDelete={handleDelete}
                    isEditMode={isEditMode}
                  />
                ))}
              </div>
            </SortableContext>
          </DndContext>
        ) : (
          <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'>
            {pluginsCopy.map((plugin) => (
              <SortablePluginCard
                key={plugin.id}
                plugin={plugin}
                onEdit={handleEdit}
                onDelete={handleDelete}
                isEditMode={isEditMode}
              />
            ))}
          </div>
        ))}

      <EditPluginDialog
        plugin={editingPlugin}
        isOpen={isEditDialogOpen}
        onOpenChange={setIsEditDialogOpen}
        onSave={handleSave}
      />

      <AddPluginDialog
        isOpen={isAddDialogOpen}
        onOpenChange={setIsAddDialogOpen}
        onSave={handleSave}
      />

      <DeleteConfirmDialog />
    </div>
  );
}
