import landing from './landing'
import common from './common'
import dashboard from './dashboard'
import admin from './admin'
import misc from './misc'
import legacyGas from './legacyGas'
import backfill from './backfill'
import { mergeMissingMessages } from '../mergeMissing'

export default mergeMissingMessages({
  ...landing,
  ...common,
  ...dashboard,
  admin,
  ...misc,
}, legacyGas, backfill)
