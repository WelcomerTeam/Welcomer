import { createApp } from "vue";
import App from "./App.vue";

import router from "./router";
import store from "./store";

import VueLazyload from "vue3-lazyload";

import "./index.scss";

import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome";
import { library } from "@fortawesome/fontawesome-svg-core";

import {
  faDiscord,
  faPatreon,
  faPaypal,
  faUnsplash,
  faYoutube,
} from "@fortawesome/free-brands-svg-icons";
library.add(faDiscord, faPatreon, faPaypal, faUnsplash, faYoutube);

import {
  faArrowsRotate as fasArrowsRotate,
  faAt as fasAt,
  faBook as fasBook,
  faBoxesPacking as fasBoxesPacking,
  faChartLine as fasChartLine,
  faCheck as fasCheck,
  faChevronDown as fasChevronDown,
  faChevronUp as fasChevronUp,
  faCircle as fasCircle,
  faCircleQuestion as fasCircleQuestion,
  faClose as fasClose,
  faCog as fasCog,
  faDoorClosed as fasDoorClosed,
  faDoorOpen as fasDoorOpen,
  faDownLong as fasDownLong,
  faFaceLaugh as fasFaceLaugh,
  faFileImage as fasFileImage,
  faGripVertical as fasGripVertical,
  faHashtag as fasHashtag,
  faHeart as fasHeart,
  faHeartPulse as fasHeartPulse,
  faImage as fasImage,
  faImages as fasImages,
  faInfo as fasInfo,
  faLifeRing as fasLifeRing,
  faListCheck as fasListCheck,
  faListOl as fasListOl,
  faMicrophoneLines as fasMicrophoneLines,
  faMoon as fasMoon,
  faPaintRoller as fasPaintRoller,
  faPenToSquare as fasPenToSquare,
  faPersonCircleQuestion as fasPersonCircleQuestion,
  faPhotoFilm as fasPhotoFilm,
  faPlugCircleBolt as fasPlugCircleBolt,
  faPlugCirclePlus as fasPlugCirclePlus,
  faPlus as fasPlus,
  faRotateRight as fasRotateRight,
  faShield as fasShield,
  faSquare as fasSquare,
  faSun as fasSun,
  faTachographDigital as fasTachographDigital,
  faToolbox as fasToolbox,
  faTriangleExclamation as fasTriangleExclamation,
  faTurnDown as fasTurnDown,
  faTurnUp as fasTurnUp,
  faUpLong as fasUpLong,
  faUser as fasUser,
  faUserCheck as fasUserCheck,
  faUserClock as fasUserClock,
  faUserGroup as fasUserGroup,
  faUserMinus as fasUserMinus,
  faUserPlus as fasUserPlus,
  faUserShield as fasUserShield,
  faWrench as fasWrench,
} from "@fortawesome/pro-solid-svg-icons";

import {
  faChartLine as farChartLine,
  faCopy as farCopy,
  faDoorClosed as farDoorClosed,
  faHeart as farHeart,
  faListCheck as farListCheck,
  faListOl as farListOl,
  faMicrophoneLines as farMicrophoneLines,
  faUserCheck as farUserCheck,
  faUserClock as farUserClock,
  faUserGroup as farUserGroup,
  faUserMinus as farUserMinus,
  faUserPlus as farUserPlus,
  faWrench as farWrench,
} from "@fortawesome/pro-regular-svg-icons";

import {
  faBadgeCheck as falBadgeCheck,
} from "@fortawesome/sharp-light-svg-icons";

import {
  faPersonDigging as fadPersonDigging
} from "@fortawesome/pro-duotone-svg-icons";

library.add(
  fasArrowsRotate,
  fasAt,
  fasBook,
  fasBoxesPacking,
  fasChartLine,
  fasCheck,
  fasChevronDown,
  fasChevronUp,
  fasCircle,
  fasCircleQuestion,
  fasClose,
  fasCog,
  fasDoorClosed,
  fasDoorOpen,
  fasDownLong,
  fasFaceLaugh,
  fasFileImage,
  fasGripVertical,
  fasHashtag,
  fasHeart,
  fasHeartPulse,
  fasImage,
  fasImages,
  fasInfo,
  fasLifeRing,
  fasListCheck,
  fasListOl,
  fasMicrophoneLines,
  fasMoon,
  fasPaintRoller,
  fasPenToSquare,
  fasPersonCircleQuestion,
  fasPhotoFilm,
  fasPlugCircleBolt,
  fasPlugCirclePlus,
  fasPlus,
  fasRotateRight,
  fasShield,
  fasSquare,
  fasSun,
  fasTachographDigital,
  fasToolbox,
  fasTriangleExclamation,
  fasTurnDown,
  fasTurnUp,
  fasUpLong,
  fasUser,
  fasUserCheck,
  fasUserClock,
  fasUserGroup,
  fasUserMinus,
  fasUserPlus,
  fasUserShield,
  fasWrench,

  farChartLine,
  farCopy,
  farDoorClosed,
  farHeart,
  farListCheck,
  farListOl,
  farMicrophoneLines,
  farUserCheck,
  farUserClock,
  farUserGroup,
  farUserMinus,
  farUserPlus,
  farWrench,

  fadPersonDigging,

  falBadgeCheck,
);

createApp(App)
  .component("font-awesome-icon", FontAwesomeIcon)
  .use(VueLazyload)
  .use(store)
  .use(router)
  .mount("#app");
