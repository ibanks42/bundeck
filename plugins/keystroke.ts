import { Key, keyboard } from '@nut-tree-fork/nut-js';

const BUNDECK_KEYS = ["104", "108", "94"];
const keys = BUNDECK_KEYS.map((k) => Number(k));

const keyNames = keys.map((k) => Key[k]);

await keyboard.type(...keys);

console.log('Pressed', keyNames.join('+'));
