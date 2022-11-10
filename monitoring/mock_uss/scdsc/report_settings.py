from monitoring.mock_uss.scdsc.muss_reports import MussReport
from monitoring.mock_uss.scdsc.muss_report_recorder import MussReportRecorder

reprt = MussReport()
reprt_recorder = MussReportRecorder(reprt)

def reset():
    global reprt
    reprt.reset()
