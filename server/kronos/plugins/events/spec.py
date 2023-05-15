from typing import Dict, List

import pluggy

from kronos.settings import settings

hookspec = pluggy.HookspecMarker(settings.FAUST_APP_NAME)  # noqa


class EventsPluginSpec(object):
    @hookspec
    def publish_event(self, payload: List[Dict]):
        """
        Publishes an event
        """
