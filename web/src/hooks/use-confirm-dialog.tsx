import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useState } from 'react';

interface UseConfirmDialogProps {
  title: string;
  description: string;
  cancelText?: string;
  confirmText?: string;
  confirmVariant?:
    | 'default'
    | 'destructive'
    | 'outline'
    | 'secondary'
    | 'ghost'
    | 'link';
}

export function useConfirmDialog({
  title,
  description,
  cancelText = 'Cancel',
  confirmText = 'Confirm',
  confirmVariant = 'default',
}: UseConfirmDialogProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [callback, setCallback] = useState<() => void>(() => () => {});

  const onConfirm = () => {
    callback();
    setIsOpen(false);
  };

  const confirm = (onConfirm: () => void) => {
    setCallback(() => onConfirm);
    setIsOpen(true);
  };

  const ConfirmDialog = () => (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant='outline' onClick={() => setIsOpen(false)}>
            {cancelText}
          </Button>
          <Button variant={confirmVariant} onClick={onConfirm}>
            {confirmText}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );

  return { confirm, ConfirmDialog };
}
