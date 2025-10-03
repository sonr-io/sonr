import { Window as KeplrWindow } from "@sonr.io/es/registry";

import { Window as CompassWindow } from "./compass/types";
import { Window as CosmostationWindow } from "./cosmostation/types";
import { Window as LeapWindow } from "./leap/types";
import { Window as EthereumWindow } from "./metamask-injective/types";
import { Window as NinjiWindow } from "./ninji/types";
import { Window as OWalletWindow } from "./owallet/types";
import { Window as StationWindow } from "./station/types";

declare global {
  interface Window
    extends KeplrWindow,
      CosmostationWindow,
      StationWindow,
      LeapWindow,
      CompassWindow,
      EthereumWindow,
      NinjiWindow,
      OWalletWindow {}
}
