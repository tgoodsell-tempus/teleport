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

import { FileStorage } from 'teleterm/services/fileStorage';

export function createMockFileStorage(opts?: {
  filePath: string;
}): FileStorage {
  let state = {};
  return {
    put(key: string, json: any) {
      state[key] = json;
    },

    get<T>(key?: string): T {
      return key ? state[key] : (state as T);
    },

    writeSync() {},

    replace(json: any) {
      state = json;
    },

    getFilePath(): string {
      return opts?.filePath || '';
    },
  };
}
