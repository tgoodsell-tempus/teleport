/**
 * Copyright 2023 Gravitational, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { Platform } from 'teleterm/mainProcess/types';
import {
  KeyboardShortcutAction,
  ConfigService,
} from 'teleterm/services/config';

import { getKeyName } from './getKeyName';
import {
  KeyboardShortcutEvent,
  KeyboardShortcutEventSubscriber,
} from './types';

export class KeyboardShortcutsService {
  private eventsSubscribers = new Set<KeyboardShortcutEventSubscriber>();
  private readonly acceleratorsToActions = new Map<
    string,
    KeyboardShortcutAction[]
  >();
  /**
   * Modifier keys must be defined in the following order:
   * Control-Option-Shift-Command for macOS
   * Ctrl-Alt-Shift for other platforms
   */
  private readonly shortcutsConfig: Record<KeyboardShortcutAction, string>;

  constructor(
    private platform: Platform,
    private configService: ConfigService
  ) {
    this.shortcutsConfig = {
      tab1: this.configService.get('keymap.tab1').value,
      tab2: this.configService.get('keymap.tab2').value,
      tab3: this.configService.get('keymap.tab3').value,
      tab4: this.configService.get('keymap.tab4').value,
      tab5: this.configService.get('keymap.tab5').value,
      tab6: this.configService.get('keymap.tab6').value,
      tab7: this.configService.get('keymap.tab7').value,
      tab8: this.configService.get('keymap.tab8').value,
      tab9: this.configService.get('keymap.tab9').value,
      closeTab: this.configService.get('keymap.closeTab').value,
      previousTab: this.configService.get('keymap.previousTab').value,
      nextTab: this.configService.get('keymap.nextTab').value,
      newTab: this.configService.get('keymap.newTab').value,
      openQuickInput: this.configService.get('keymap.openQuickInput').value,
      openConnections: this.configService.get('keymap.openConnections').value,
      openClusters: this.configService.get('keymap.openClusters').value,
      openProfiles: this.configService.get('keymap.openProfiles').value,
    };
    this.acceleratorsToActions = mapAcceleratorsToActions(this.shortcutsConfig);
    this.attachKeydownHandler();
  }

  subscribeToEvents(subscriber: KeyboardShortcutEventSubscriber): void {
    this.eventsSubscribers.add(subscriber);
  }

  unsubscribeFromEvents(subscriber: KeyboardShortcutEventSubscriber): void {
    this.eventsSubscribers.delete(subscriber);
  }

  getShortcutsConfig() {
    return this.shortcutsConfig;
  }

  /**
   * Some actions can get assigned the same accelerators.
   * This method returns them.
   */
  getDuplicateAccelerators(): Record<string, KeyboardShortcutAction[]> {
    return Array.from(this.acceleratorsToActions.entries())
      .filter(([, shortcuts]) => shortcuts.length > 1)
      .reduce<Record<string, KeyboardShortcutAction[]>>(
        (accumulator, [accelerator, actions]) => {
          accumulator[accelerator] = actions;
          return accumulator;
        },
        {}
      );
  }

  private attachKeydownHandler(): void {
    const handleKeydown = (event: KeyboardEvent): void => {
      const shortcutAction = this.getShortcutAction(event);
      if (!shortcutAction) {
        return;
      }

      event.preventDefault();
      event.stopPropagation();
      this.notifyEventsSubscribers({ action: shortcutAction });
    };

    window.addEventListener('keydown', handleKeydown, {
      capture: true,
    });
  }

  private getShortcutAction(
    event: KeyboardEvent
  ): KeyboardShortcutAction | undefined {
    // If only a modifier is pressed, `code` is this modifier name
    // (in case of a combination like "Cmd+A", `code` is "A").
    // We do not support modifier-only accelerators, so we can skip the further checks.
    if (
      event.code.includes('Shift') ||
      event.code.includes('Meta') ||
      event.code.includes('Alt') ||
      event.code.includes('Control')
    ) {
      return;
    }
    const accelerator = [
      ...this.getPlatformModifierKeys(event),
      getKeyName(event),
    ]
      .filter(Boolean)
      .join('+');

    // always return the first action (in case of duplicate accelerators)
    return this.acceleratorsToActions.get(accelerator)?.[0];
  }

  /**
   * It is important that these modifiers are in the same order as in `getKeyboardShortcutSchema#getSupportedModifiers`.
   * Consider creating "one source of truth" for them.
   */
  private getPlatformModifierKeys(event: KeyboardEvent): string[] {
    switch (this.platform) {
      case 'darwin':
        return [
          event.ctrlKey && 'Control',
          event.altKey && 'Option',
          event.shiftKey && 'Shift',
          event.metaKey && 'Command',
        ];
      default:
        return [
          event.ctrlKey && 'Ctrl',
          event.altKey && 'Alt',
          event.shiftKey && 'Shift',
        ];
    }
  }

  private notifyEventsSubscribers(event: KeyboardShortcutEvent): void {
    this.eventsSubscribers.forEach(subscriber => subscriber(event));
  }
}

/** Inverts shortcuts-keys pairs to allow accessing shortcut by an accelerator. */
function mapAcceleratorsToActions(
  shortcutsConfig: Record<KeyboardShortcutAction, string>
): Map<string, KeyboardShortcutAction[]> {
  const acceleratorsToActions = new Map<string, KeyboardShortcutAction[]>();
  Object.entries(shortcutsConfig).forEach(([action, accelerator]) => {
    // empty accelerator means that an empty string was provided in the config file, so the shortcut is disabled.
    if (!accelerator) {
      return;
    }
    acceleratorsToActions.set(accelerator, [
      ...(acceleratorsToActions.get(accelerator) || []),
      action as KeyboardShortcutAction,
    ]);
  });
  return acceleratorsToActions;
}
