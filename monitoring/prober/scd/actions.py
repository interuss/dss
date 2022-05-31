from monitoring.monitorlib.infrastructure import UTMClientSession
from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import SCOPE_CM, SCOPE_SC, SCOPE_CI, SCOPE_CP


def _read_both_scope(scd_api: str) -> str:
    if scd_api == scd.API_0_3_17:
        return '{} {}'.format(SCOPE_SC, SCOPE_CP)
    else:
        raise NotImplementedError('Unsupported API version {}'.format(scd_api))


def delete_constraint_reference_if_exists(id: str, scd_session: UTMClientSession, scd_api: str):
    resp = scd_session.get('/constraint_references/{}'.format(id), scope=SCOPE_CM)
    if resp.status_code == 200:
        if scd_api == scd.API_0_3_17:
            existing_constraint = resp.json().get('constraint_reference', None)
            resp = scd_session.delete('/constraint_references/{}/{}'.format(id, existing_constraint['ovn']), scope=SCOPE_CM)
        else:
            raise NotImplementedError('Unsupported API version {}'.format(scd_api))
        assert resp.status_code == 200, '{}: {}'.format(resp.url, resp.content)
    elif resp.status_code == 404:
        # As expected.
        pass
    else:
        assert False, resp.content


def delete_subscription_if_exists(sub_id: str, scd_session: UTMClientSession, scd_api: str):
    resp = scd_session.get('/subscriptions/{}'.format(sub_id), scope=SCOPE_SC)
    if resp.status_code == 200:
        if scd_api == scd.API_0_3_17:
            sub = resp.json().get('subscription', None)
            resp = scd_session.delete('/subscriptions/{}/{}'.format(sub_id, sub['version']), scope=_read_both_scope(scd_api))
        else:
            raise NotImplementedError('Unsupported API version {}'.format(scd_api))
        assert resp.status_code == 200, resp.content
    elif resp.status_code == 404:
        # As expected.
        pass
    else:
        assert False, resp.content


def delete_operation_if_exists(id: str, scd_session: UTMClientSession, scd_api: str):
    if scd_api == scd.API_0_3_17:
        url = '/operational_intent_references/{}'
    else:
        assert False, 'Unsupported API {}'.format(scd_api)
    resp = scd_session.get(url.format(id), scope=SCOPE_SC)
    if resp.status_code == 200:
        if scd_api == scd.API_0_3_17:
            ovn = resp.json()['operational_intent_reference']['ovn']
            resp = scd_session.delete('/operational_intent_references/{}/{}'.format(id, ovn))
        assert resp.status_code == 200, resp.content
    elif resp.status_code == 404:
        # As expected.
        pass
    else:
        assert False, 'Unsupported API {}'.format(scd_api)
