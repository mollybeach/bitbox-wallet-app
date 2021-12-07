/**
 * Copyright 2018 Shift Devices AG
 * Copyright 2021 Shift Crypto AG
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

import { Component} from 'react';
import { apiGet } from '../../utils/request';
import * as style from './qrcode.module.css';

interface Props {
    data?: string;
    size?: number;
}

interface State {
    src: string;
}

class QRCode extends Component<Props, State> {
    public readonly state: State = {
        src: '',
    }

    public static defaultProps = {
        size: 256,
    };

    public componentDidMount() {
        this.update(this.props.data);
    }

    public componentWillReceiveProps({ data }) {
        if (this.props.data !== data) {
            this.update(data);
        }
    }

    private update = (data: string | undefined) => {
        this.setState({ src: '' });
        if (data !== undefined) {
            apiGet('qr?data=' + encodeURIComponent(data)).then(src => this.setState({ src }));
        }
    }

    public render() {
        const { data, size } = this.props;
        const { src } = this.state;
        if (!src) {
            if (data !== undefined) {
                return <div className={style.empty}></div>;
            }
            return (
                <svg className={style.empty} style={{verticalAlign: 'middle',fill: 'currentColor',overflow: 'hidden'}} viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg">
                    <path d="M433.282171 47.939849 94.944396 47.939849c-23.327288 0-42.304502 18.975167-42.304502 42.300408l0 338.341868c0 23.324218 18.977213 42.300408 42.304502 42.300408L433.282171 470.882534c23.324218 0 42.300408-18.97619 42.300408-42.300408L475.582579 90.240257C475.581556 66.915016 456.606389 47.939849 433.282171 47.939849zM438.742533 428.582125c0 3.01159-2.449795 5.461385-5.461385 5.461385L94.944396 434.04351c-3.013637 0-5.465478-2.450818-5.465478-5.461385L89.478917 90.240257c0-3.01159 2.451842-5.461385 5.465478-5.461385L433.282171 84.778872c3.01159 0 5.461385 2.449795 5.461385 5.461385L438.743556 428.582125zM322.914504 173.887352l-117.608581 0c-14.733563 0-26.719548 11.985986-26.719548 26.719548l0 117.608581c0 14.735609 11.985986 26.724665 26.719548 26.724665l117.608581 0c14.735609 0 26.724665-11.989055 26.724665-26.724665l0-117.608581C349.640192 185.873338 337.651137 173.887352 322.914504 173.887352zM312.801169 308.101123l-97.374748 0 0-97.374748 97.374748 0L312.801169 308.101123zM322.914504 675.606195l-117.608581 0c-14.733563 0-26.719548 11.989055-26.719548 26.724665l0 117.608581c0 14.732539 11.985986 26.719548 26.719548 26.719548l117.608581 0c14.735609 0 26.724665-11.987009 26.724665-26.719548l0-117.608581C349.640192 687.59525 337.651137 675.606195 322.914504 675.606195zM312.801169 809.819966l-97.374748 0 0-97.374748 97.374748 0L312.801169 809.819966zM699.123818 344.940147l117.608581 0c14.735609 0 26.724665-11.989055 26.724665-26.724665l0-117.608581c0-14.733563-11.989055-26.719548-26.724665-26.719548l-117.608581 0c-14.732539 0-26.719548 11.985986-26.719548 26.719548l0 117.608581C672.40427 332.951091 684.391279 344.940147 699.123818 344.940147zM709.243293 210.726376l97.374748 0 0 97.374748-97.374748 0L709.243293 210.726376zM433.282171 546.435277 94.944396 546.435277c-23.327288 0-42.304502 18.97619-42.304502 42.300408L52.639894 927.077553c0 23.325241 18.977213 42.301432 42.304502 42.301432L433.282171 969.378985c23.324218 0 42.300408-18.97619 42.300408-42.301432L475.582579 588.735685C475.581556 565.411467 456.606389 546.435277 433.282171 546.435277zM438.742533 927.077553c0 3.01159-2.449795 5.462408-5.461385 5.462408L94.944396 932.539962c-3.013637 0-5.465478-2.450818-5.465478-5.462408L89.478917 588.735685c0-3.010567 2.451842-5.461385 5.465478-5.461385L433.282171 583.2743c3.01159 0 5.461385 2.450818 5.461385 5.461385L438.743556 927.077553zM590.849836 547.776831c-23.325241 0-42.301432 18.978237-42.301432 42.305525l0 194.649212c0 10.172687 8.246825 18.419512 18.419512 18.419512s18.419512-8.246825 18.419512-18.419512L585.387427 590.082356c0-3.013637 2.450818-5.466502 5.462408-5.466502l204.825992 0c10.172687 0 18.419512-8.246825 18.419512-18.419512s-8.246825-18.419512-18.419512-18.419512L590.849836 547.776831zM588.757175 470.882534l338.341868 0c23.325241 0 42.301432-18.97619 42.301432-42.300408L969.400474 90.240257c0-23.324218-18.97619-42.300408-42.301432-42.300408L588.757175 47.939849c-23.324218 0-42.300408 18.975167-42.300408 42.300408l0 338.341868C546.456766 451.906343 565.432956 470.882534 588.757175 470.882534zM583.295789 90.240257c0-3.01159 2.450818-5.461385 5.461385-5.461385l338.341868 0c3.01159 0 5.462408 2.449795 5.462408 5.461385l0 338.341868c0 3.01159-2.450818 5.461385-5.462408 5.461385L588.757175 434.04351c-3.010567 0-5.461385-2.450818-5.461385-5.461385L583.295789 90.240257zM969.676767 566.196343c0-10.172687-8.246825-18.419512-18.419512-18.419512l-67.50444 0c-10.172687 0-18.419512 8.246825-18.419512 18.419512s8.246825 18.419512 18.419512 18.419512l49.084928 0 0 71.009264L779.708158 655.625118c-10.172687 0-18.419512 8.246825-18.419512 18.419512s8.246825 18.419512 18.419512 18.419512l171.549098 0c10.172687 0 18.419512-8.246825 18.419512-18.419512 0-0.076748-0.010233-0.151449-0.011256-0.228197 0.001023-0.076748 0.011256-0.150426 0.011256-0.228197L969.676767 566.196343zM630.202099 933.142689l-44.813648 0 0-51.2922c0-10.172687-8.246825-18.419512-18.419512-18.419512s-18.419512 8.246825-18.419512 18.419512l0 69.711711c0 10.172687 8.246825 18.419512 18.419512 18.419512l63.23316 0c10.172687 0 18.419512-8.246825 18.419512-18.419512S640.374786 933.142689 630.202099 933.142689zM834.04981 790.359752c0-10.172687-8.246825-18.419512-18.419512-18.419512L699.257871 771.940241c-10.172687 0-18.419512 8.246825-18.419512 18.419512l0 163.013701c0 10.172687 8.246825 18.419512 18.419512 18.419512s18.419512-8.246825 18.419512-18.419512L717.677383 808.779264 815.630299 808.779264C825.802986 808.779264 834.04981 800.532439 834.04981 790.359752zM951.060781 723.425294c-10.172687 0-18.419512 8.246825-18.419512 18.419512l0 194.301288L765.943666 936.146093c-10.172687 0-18.419512 8.246825-18.419512 18.419512s8.246825 18.419512 18.419512 18.419512l185.117115 0c10.172687 0 18.419512-8.246825 18.419512-18.419512L969.480292 741.844805C969.480292 731.672118 961.234491 723.425294 951.060781 723.425294zM655.979182 623.620194c-10.172687 0-18.419512 8.246825-18.419512 18.419512l0 59.869552c0 10.172687 8.246825 18.419512 18.419512 18.419512s18.419512-8.246825 18.419512-18.419512l0-59.869552C674.398694 631.867018 666.151869 623.620194 655.979182 623.620194z" />
                </svg>
            );
        }
        return (
            <img
                width={size}
                height={size}
                src={src}
            />
        );
    }
}

export { QRCode };
