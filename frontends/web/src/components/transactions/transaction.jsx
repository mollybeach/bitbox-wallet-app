import { Component } from 'preact';
import { translate } from 'react-i18next';
import style from './transaction.css';
import IN from './assets/icon-transfer-in.svg';
import OUT from './assets/icon-transfer-out.svg';
// TODO: import SELF from './assets/icon-transfer-self.svg';

const transferIconMap = {
    receive: IN,
    send_to_self: IN,
    send: OUT
};

@translate()
export default class Transaction extends Component {
    state = {
        collapsed: true,
    }

    onUncollapse = e => {
        this.setState(({ collapsed }) => ({ collapsed: !collapsed }));
    }

    parseTime = (time) => {
        let arr;
        const months = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];
        const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
        arr = time.split(' ');
        arr.pop();
        const dt = new Date(Date.parse(arr.join(' ')));
        return `${days[dt.getDay()]} ${dt.getDate()} ${months[dt.getMonth()]} ${dt.getFullYear()} - ${dt.getHours()}:${dt.getMinutes()}`;
    }

    render({
        t,
        explorerURL,
        type,
        id,
        amount,
        fiat = '',
        fiat_historical = '',
        fee,
        height,
        time,
        addresses,
    }, {
        collapsed,
    }) {
        const badge = t(`transaction.badge.${type}`);
        const sign = ((type === 'send') && '−') || ((type === 'receive') && '+') || null;
        // TODO: check if 'Time not yet available' is needed
        const date = time ? this.parseTime(time) : (height <= 0 ? t('transaction.pending') : 'Time not yet available');
        return (
            <div
                className={[style.transactionContainer, collapsed ? style.collapsed : style.expanded].join(' ')}
                onClick={this.onUncollapse}>
                <div className={['flex', 'flex-row', 'flex-start', 'flex-items-start', style.transaction].join(' ')}>
                    {/*
                    <div>
                        <img src={transferIconMap[type]} height="22" style="margin-right: var(--spacing-default)" />
                    </div>
                    */}
                    <div className="flex-1">
                        <div className={['flex', 'flex-row', 'flex-between', 'flex-items-start', style.row].join(' ')}>
                            <div className="flex flex-row flex-start flex-items-center">
                                <div className={[style.transactionLabel, style[type]].join(' ')}>
                                    {badge}
                                </div>
                                <div className={style.address}>{addresses}</div>
                            </div>
                            <div className={[style.amount, style[type]].join(' ')}>
                                {sign}{amount}
                            </div>
                        </div>
                        <div className={['flex', 'flex-row', 'flex-between', 'flex-items-start', style.row].join(' ')}>
                            <div className={style.date}>{date}</div>
                            <div className={[style.amount, style.converted].join(' ')}>{fiat}</div>
                        </div>

                        <div hidden={collapsed ? 'hidden' : null} className={style.collapsedContent}>
                            <div className={['flex', 'flex-row', 'flex-start', 'flex-items-start', style.row].join(' ')}>
                                <div>
                                    <div className={style.transactionLabel}>{t('transaction.confirmation')}</div>
                                    <div className={style.address}>{height}</div>
                                </div>
                                {
                                    fee && (
                                        <div>
                                            <div className={style.transactionLabel}>{t('transaction.fee')}</div>
                                            <div className={style.address}>{fee}</div>
                                        </div>
                                    )
                                }
                                {
                                    fiat_historical && (
                                        <div style="align-self: flex-end; margin-left: auto; text-align: right;">
                                            <div className={style.transactionLabel} style="margin-right: 0;">
                                              {t('transaction.fiatHistorical')}
                                            </div>
                                            <div className={style.address}>{fiat_historical}</div>
                                        </div>
                                    )
                                }
                            </div>
                            <div className={style.row}>
                                <div className={style.transactionLabel}>{t('transaction.explorer')}</div>
                                <div className={style.address}>
                                    <a href={ explorerURL + id } target="_blank">{id}</a>
                                </div>
                            </div>
                        </div>

                    </div>
                </div>
            </div>
        );
    }
}
