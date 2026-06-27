import { defineConfigWithVueTs, vueTsConfigs } from '@vue/eslint-config-typescript'
import pluginVue from 'eslint-plugin-vue'
import { globalIgnores } from 'eslint/config'

export default defineConfigWithVueTs(
  { name: 'app/files-to-lint', files: ['**/*.{vue,ts}'] },

  globalIgnores(['**/dist/**']),

  ...pluginVue.configs['flat/essential'],
  vueTsConfigs.recommended,
)
