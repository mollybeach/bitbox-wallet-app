import { h } from 'preact';
import i18n from '../../../i18n/i18n';
import { ButtonLink } from '../../../components/forms';
import { Guide } from '../../../components/guide/guide';
import Backups from '../../../components/backups/backups';

export default function ManageBackups({
    deviceID,
    guide,
}) {
    return (
        <div class="contentWithGuide">
            <div class="container">
                <div class="headerContainer">
                    <div class="header">
                        <h2>{i18n.t('backup.title')}</h2>
                    </div>
                </div>
                <Backups
                    deviceID={deviceID}
                    showCreate={true}>
                    <ButtonLink
                        secondary
                        href={`/device/${deviceID}`}
                        style="position: absolute; right: 0;">
                        {i18n.t('button.back')}
                    </ButtonLink>
                </Backups>
            </div>
            <Guide guide={guide} screen="backups" />
        </div>
    );
}
