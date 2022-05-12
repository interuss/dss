import functools

from monitoring.deployment_manager.infrastructure import Context


def requires_v1_dss(func):
    """Make sure v1 DSS is well-specified"""
    @functools.wraps(func)
    def wrapper_v1_dss_required(context: Context, *args, **kwargs):
        if 'dss' not in context.spec:
            raise ValueError('DSS system is not defined in deployment configuration')
        return func(context, *args, **kwargs)
    return wrapper_v1_dss_required
