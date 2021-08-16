import json
from flask_wtf import FlaskForm
from wtforms import StringField, SubmitField, TextAreaField, BooleanField, IntegerField, widgets, SelectMultipleField
from wtforms.validators import DataRequired, ValidationError
from wtforms.widgets import HiddenInput


class UserConfig(FlaskForm):
    auth_spec = StringField('Auth Spec', validators=[DataRequired()])
    user_config = TextAreaField('User Config', validators=[DataRequired()])
    sample_report = BooleanField('Sample Report')
    file_count = IntegerField(widget=HiddenInput())
    submit = SubmitField('Submit')

    def validate_user_config(form, field):
        user_config = json.loads(field.data)
        expected_keys = {'injection_targets', 'observers'}
        if not expected_keys.issubset(set(user_config)):
            message = f'missing fields in config object {expected_keys - set(user_config)}'
            raise ValidationError(message)
        if int(form.file_count.data) < len(user_config['injection_targets']):
            raise ValidationError(
                'Not enough flight states files provided for each injection_targets.')

class MultiCheckboxField(SelectMultipleField):
    widget = widgets.ListWidget(prefix_label=False)
    option_widget = widgets.CheckboxInput()

class SimpleForm(FlaskForm):
    # string_of_files = ['one\r\ntwo\r\nthree\r\n']
    # list_of_files = string_of_files[0].split()
    # create a list of value/description tuples
    # files = StringField(widget=HiddenInput())
    # files_list = str(files.data).split(',')
    # files = [(x, x) for x in files_list]
    example = MultiCheckboxField('Label', choices=[])

class TestsExecuteForm(FlaskForm):
    flight_records = MultiCheckboxField('Flight Records', choices=[], validators=[DataRequired()])
    auth_spec = StringField('Auth Spec', validators=[DataRequired()])
    user_config = TextAreaField('User Config', validators=[DataRequired()])
    sample_report = BooleanField('Sample Report')
    submit = SubmitField('Submit')

    def validate_user_config(form, field):
        user_config = json.loads(field.data)
        expected_keys = {'injection_targets', 'observers'}
        if not expected_keys.issubset(set(user_config)):
            message = f'missing fields in config object {expected_keys - set(user_config)}'
            raise ValidationError(message)
        if len(form.flight_records.data) < len(user_config['injection_targets']):
            raise ValidationError(
                'Not enough flight states files provided for each injection_targets.')
    # def validate_flight_records(form, field):
