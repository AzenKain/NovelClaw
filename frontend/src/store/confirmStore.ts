import { create } from 'zustand';

export type ConfirmVariant = 'danger' | 'warning' | 'info' | 'success';

export interface ConfirmModalOptions {
  title?: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: ConfirmVariant;
  icon?: 'trash' | 'alert' | 'warning' | 'info';
  isAlert?: boolean; // If true, single OK button (acts as Alert modal)
}

interface ConfirmState {
  isOpen: boolean;
  options: ConfirmModalOptions;
  resolve: ((value: boolean) => void) | null;
  open: (options: ConfirmModalOptions) => Promise<boolean>;
  close: (confirmed: boolean) => void;
}

export const useConfirmStore = create<ConfirmState>((set, get) => ({
  isOpen: false,
  options: { message: '' },
  resolve: null,
  open: (options) => {
    return new Promise<boolean>((resolve) => {
      set({
        isOpen: true,
        options,
        resolve,
      });
    });
  },
  close: (confirmed) => {
    const { resolve } = get();
    if (resolve) {
      resolve(confirmed);
    }
    set({
      isOpen: false,
      resolve: null,
    });
  },
}));

/**
 * Global helper to show an interactive Confirm Modal dialog anywhere.
 * Replaces native browser window.confirm().
 */
export const showConfirmModal = (options: ConfirmModalOptions | string): Promise<boolean> => {
  const opts = typeof options === 'string' ? { message: options } : options;
  return useConfirmStore.getState().open(opts);
};

/**
 * Global helper to show an interactive Alert Modal dialog anywhere.
 * Replaces native browser window.alert().
 */
export const showAlertModal = (options: Omit<ConfirmModalOptions, 'isAlert'> | string): Promise<void> => {
  const opts = typeof options === 'string' ? { message: options, isAlert: true } : { ...options, isAlert: true };
  return useConfirmStore.getState().open(opts).then(() => {});
};
